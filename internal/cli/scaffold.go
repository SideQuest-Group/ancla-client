package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// pythonProject holds the parsed metadata from a pyproject.toml file.
type pythonProject struct {
	Framework      string // litestar, fastapi, flask, django, or ""
	Entrypoint     string // e.g. "myapp:main" from [project.scripts]
	PythonVersion  string // e.g. "3.14" from requires-python
	PackageManager string // "uv" or "pip"
	ProjectName    string // from [project].name
}

// detectPython checks for pyproject.toml in the current directory and parses
// it for framework, entrypoint, Python version, and package manager.
// Returns nil if no pyproject.toml is found.
func detectPython() *pythonProject {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	pyprojectPath := filepath.Join(cwd, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); os.IsNotExist(err) {
		return nil
	}

	var pyproject struct {
		Project struct {
			Name           string            `toml:"name"`
			RequiresPython string            `toml:"requires-python"`
			Dependencies   []string          `toml:"dependencies"`
			Scripts        map[string]string `toml:"scripts"`
		} `toml:"project"`
	}

	if _, err := toml.DecodeFile(pyprojectPath, &pyproject); err != nil {
		return nil
	}

	p := &pythonProject{
		ProjectName: pyproject.Project.Name,
	}

	// Detect framework from dependencies
	for _, dep := range pyproject.Project.Dependencies {
		depLower := strings.ToLower(dep)
		// Strip version specifiers for matching
		name := strings.FieldsFunc(depLower, func(r rune) bool {
			return r == '>' || r == '<' || r == '=' || r == '!' || r == '~' || r == '[' || r == ';'
		})[0]
		name = strings.TrimSpace(name)
		switch {
		case name == "litestar":
			p.Framework = "litestar"
		case name == "fastapi":
			p.Framework = "fastapi"
		case name == "flask":
			p.Framework = "flask"
		case name == "django":
			p.Framework = "django"
		}
	}

	// Detect entrypoint from [project.scripts]
	if len(pyproject.Project.Scripts) > 0 {
		for _, v := range pyproject.Project.Scripts {
			p.Entrypoint = v
			break
		}
	}

	// Detect Python version from requires-python
	p.PythonVersion = parsePythonVersion(pyproject.Project.RequiresPython)

	// Detect package manager: uv.lock → uv, else pip
	if _, err := os.Stat(filepath.Join(cwd, "uv.lock")); err == nil {
		p.PackageManager = "uv"
	} else {
		p.PackageManager = "pip"
	}

	return p
}

// parsePythonVersion extracts a usable Docker tag version from requires-python.
// Examples: ">=3.14" → "3.14", ">=3.12,<4" → "3.12", "==3.13.*" → "3.13"
func parsePythonVersion(spec string) string {
	if spec == "" {
		return "3.14"
	}
	// Take the first version constraint
	parts := strings.Split(spec, ",")
	v := strings.TrimSpace(parts[0])
	// Strip operators
	v = strings.TrimLeft(v, "><=!~")
	v = strings.TrimSpace(v)
	// Strip wildcard
	v = strings.TrimSuffix(v, ".*")
	if v == "" {
		return "3.14"
	}
	return v
}

// frameworkCMD returns the default CMD for a given framework.
func frameworkCMD(p *pythonProject) string {
	switch p.Framework {
	case "litestar":
		return `["litestar", "run", "--host", "0.0.0.0", "--port", "8000"]`
	case "fastapi":
		return `["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]`
	case "flask":
		return `["gunicorn", "--bind", "0.0.0.0:8000", "app:app"]`
	case "django":
		return `["gunicorn", "--bind", "0.0.0.0:8000", "config.wsgi:application"]`
	default:
		if p.Entrypoint != "" {
			return fmt.Sprintf(`["python", "-m", "%s"]`, strings.Split(p.Entrypoint, ":")[0])
		}
		return `["python", "-m", "app"]`
	}
}

// frameworkProcfileCmd returns the Procfile web process command.
func frameworkProcfileCmd(p *pythonProject) string {
	switch p.Framework {
	case "litestar":
		return "litestar run --host 0.0.0.0 --port 8000"
	case "fastapi":
		return "uvicorn app:app --host 0.0.0.0 --port 8000"
	case "flask":
		return "gunicorn --bind 0.0.0.0:8000 app:app"
	case "django":
		return "gunicorn --bind 0.0.0.0:8000 config.wsgi:application"
	default:
		if p.Entrypoint != "" {
			return "python -m " + strings.Split(p.Entrypoint, ":")[0]
		}
		return "python -m app"
	}
}

// generateDockerfile generates the Dockerfile.ancla content for a Python project.
func generateDockerfile(p *pythonProject) string {
	pyVer := p.PythonVersion
	cmd := frameworkCMD(p)

	if p.PackageManager == "uv" {
		return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy

# Install dependencies
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev --no-install-project

# Copy application
COPY . .
RUN uv sync --frozen --no-dev

# Run
EXPOSE 8000
CMD %s
`, pyVer, cmd)
	}

	// pip variant
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

# Install dependencies
COPY pyproject.toml ./
RUN pip install --no-cache-dir .

# Copy application
COPY . .
RUN pip install --no-cache-dir .

# Run
EXPOSE 8000
CMD %s
`, pyVer, cmd)
}

// generateProcfile generates the Procfile.ancla content for a Python project.
func generateProcfile(p *pythonProject) string {
	return "web: " + frameworkProcfileCmd(p) + "\n"
}

// ensureDockerfile checks for Dockerfile.ancla or Dockerfile in the working
// directory. If neither exists and a Python project is detected, it offers
// to generate Dockerfile.ancla and Procfile.ancla.
func ensureDockerfile() error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Already has a Dockerfile — skip
	for _, name := range []string{"Dockerfile.ancla", "Dockerfile"} {
		if _, err := os.Stat(filepath.Join(cwd, name)); err == nil {
			return nil
		}
	}

	// Detect Python project
	p := detectPython()
	if p == nil {
		if !isQuiet() {
			fmt.Println("\n→ No Dockerfile found. No pyproject.toml detected — skipping scaffold.")
			fmt.Println("  Create a Dockerfile or Dockerfile.ancla manually before deploying.")
		}
		return nil
	}

	framework := p.Framework
	if framework == "" {
		framework = "Python"
	}

	if !isQuiet() {
		fmt.Println()
		fmt.Printf("→ No Dockerfile.ancla found.\n")
		fmt.Printf("  Detected: %s (pyproject.toml", framework)
		if p.PackageManager == "uv" {
			fmt.Print(", uv")
		}
		fmt.Println(")")
	}

	if !promptConfirm("  Generate Dockerfile.ancla?") {
		return nil
	}

	// Write Dockerfile.ancla
	dockerfilePath := filepath.Join(cwd, "Dockerfile.ancla")
	if err := os.WriteFile(dockerfilePath, []byte(generateDockerfile(p)), 0o644); err != nil {
		return fmt.Errorf("writing Dockerfile.ancla: %w", err)
	}

	// Write Procfile.ancla
	procfilePath := filepath.Join(cwd, "Procfile.ancla")
	if err := os.WriteFile(procfilePath, []byte(generateProcfile(p)), 0o644); err != nil {
		return fmt.Errorf("writing Procfile.ancla: %w", err)
	}

	fmt.Println("  ✓ Generated Dockerfile.ancla + Procfile.ancla")
	return nil
}
