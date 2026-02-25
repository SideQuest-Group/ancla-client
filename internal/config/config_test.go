package config

import (
	"os"
	"path/filepath"
	"testing"
)

// resolveSymlinks resolves symlinks in a path to handle macOS /var -> /private/var.
func resolveSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func TestLoad_Defaults(t *testing.T) {
	// Use a temp dir as HOME so no real config files are picked up.
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Clear any ANCLA_ env vars that might interfere.
	t.Setenv("ANCLA_SERVER", "")
	t.Setenv("ANCLA_API_KEY", "")

	// Change to a directory without .ancla/ to avoid local config.
	origDir, _ := os.Getwd()
	os.Chdir(tmpHome)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server != "https://ancla.dev" {
		t.Errorf("Server = %q, want %q", cfg.Server, "https://ancla.dev")
	}
	if cfg.APIKey != "" {
		t.Errorf("APIKey = %q, want empty", cfg.APIKey)
	}
}

func TestLoad_GlobalConfigFile(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	t.Setenv("ANCLA_SERVER", "")
	t.Setenv("ANCLA_API_KEY", "")

	// Create global config file.
	configDir := filepath.Join(tmpHome, ".ancla")
	os.MkdirAll(configDir, 0o755)
	configContent := []byte("server: https://custom.example.com\napi_key: global-key-123\nusername: testuser\n")
	os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpHome)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server != "https://custom.example.com" {
		t.Errorf("Server = %q, want %q", cfg.Server, "https://custom.example.com")
	}
	if cfg.APIKey != "global-key-123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "global-key-123")
	}
	if cfg.Username != "testuser" {
		t.Errorf("Username = %q, want %q", cfg.Username, "testuser")
	}
}

func TestLoad_EnvVarOverridesFile(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create global config with one server value.
	configDir := filepath.Join(tmpHome, ".ancla")
	os.MkdirAll(configDir, 0o755)
	configContent := []byte("server: https://from-file.example.com\napi_key: file-key\n")
	os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0o644)

	// Set env vars — should override file values.
	t.Setenv("ANCLA_SERVER", "https://from-env.example.com")
	t.Setenv("ANCLA_API_KEY", "env-key-456")

	origDir, _ := os.Getwd()
	os.Chdir(tmpHome)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server != "https://from-env.example.com" {
		t.Errorf("Server = %q, want %q (env should override file)", cfg.Server, "https://from-env.example.com")
	}
	if cfg.APIKey != "env-key-456" {
		t.Errorf("APIKey = %q, want %q (env should override file)", cfg.APIKey, "env-key-456")
	}
}

func TestLoad_LocalConfigOverridesGlobal(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	t.Setenv("ANCLA_SERVER", "")
	t.Setenv("ANCLA_API_KEY", "")

	// Create global config.
	globalDir := filepath.Join(tmpHome, ".ancla")
	os.MkdirAll(globalDir, 0o755)
	globalContent := []byte("server: https://global.example.com\napi_key: global-key\n")
	os.WriteFile(filepath.Join(globalDir, "config.yaml"), globalContent, 0o644)

	// Create a project directory with local .ancla/config.yaml.
	projectDir := filepath.Join(tmpHome, "projects", "myapp")
	localConfigDir := filepath.Join(projectDir, ".ancla")
	os.MkdirAll(localConfigDir, 0o755)
	localContent := []byte("org: my-org\nproject: my-project\napp: my-app\n")
	os.WriteFile(filepath.Join(localConfigDir, "config.yaml"), localContent, 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Global values should still be present.
	if cfg.Server != "https://global.example.com" {
		t.Errorf("Server = %q, want %q", cfg.Server, "https://global.example.com")
	}
	if cfg.APIKey != "global-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "global-key")
	}
	// Local values should be merged in.
	if cfg.Org != "my-org" {
		t.Errorf("Org = %q, want %q", cfg.Org, "my-org")
	}
	if cfg.Project != "my-project" {
		t.Errorf("Project = %q, want %q", cfg.Project, "my-project")
	}
	if cfg.App != "my-app" {
		t.Errorf("App = %q, want %q", cfg.App, "my-app")
	}
}

func TestFindLocalConfigDir_WalksUp(t *testing.T) {
	tmpDir := resolveSymlinks(t, t.TempDir())

	// Create .ancla/ at top level.
	anclaDir := filepath.Join(tmpDir, ".ancla")
	os.MkdirAll(anclaDir, 0o755)
	os.WriteFile(filepath.Join(anclaDir, "config.yaml"), []byte("org: test\n"), 0o644)

	// Create a nested subdirectory.
	nested := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(nested, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(nested)
	defer os.Chdir(origDir)

	got := findLocalConfigDir()
	if got != anclaDir {
		t.Errorf("findLocalConfigDir() = %q, want %q", got, anclaDir)
	}
}

func TestFindLocalConfigDir_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// No .ancla/ anywhere under tmpDir. Chdir into it so walk-up
	// will not find any .ancla directory before reaching root.
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	got := findLocalConfigDir()
	// It may or may not find one depending on real filesystem; just
	// verify it doesn't find one inside our tmpDir.
	if got != "" {
		// If it found something outside tmpDir, that's fine (real system config).
		// But it should not have found one in tmpDir.
		rel, err := filepath.Rel(tmpDir, got)
		if err == nil && !filepath.IsAbs(rel) && rel[0] != '.' {
			t.Errorf("found .ancla in tmpDir unexpectedly: %q", got)
		}
	}
}

func TestFilePath_UsesLocalIfPresent(t *testing.T) {
	tmpDir := resolveSymlinks(t, t.TempDir())

	// Create local .ancla/.
	localDir := filepath.Join(tmpDir, ".ancla")
	os.MkdirAll(localDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	got := FilePath()
	want := filepath.Join(localDir, "config.yaml")
	if got != want {
		t.Errorf("FilePath() = %q, want %q", got, want)
	}
}

func TestPaths_ReturnsGlobalAndLocal(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	tmpDir := resolveSymlinks(t, t.TempDir())
	localDir := filepath.Join(tmpDir, ".ancla")
	os.MkdirAll(localDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	global, local := Paths()
	wantGlobal := filepath.Join(tmpHome, ".ancla", "config.yaml")
	wantLocal := filepath.Join(localDir, "config.yaml")

	if global != wantGlobal {
		t.Errorf("global = %q, want %q", global, wantGlobal)
	}
	if local != wantLocal {
		t.Errorf("local = %q, want %q", local, wantLocal)
	}
}

func TestSave_CreatesConfigFile(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		Server:   "https://saved.example.com",
		APIKey:   "saved-key",
		Username: "saveduser",
		Email:    "saved@example.com",
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file was created.
	path := filepath.Join(tmpHome, ".ancla", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", path)
	}

	// Reload and verify values.
	origDir, _ := os.Getwd()
	os.Chdir(tmpHome)
	defer os.Chdir(origDir)

	t.Setenv("ANCLA_SERVER", "")
	t.Setenv("ANCLA_API_KEY", "")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if loaded.Server != "https://saved.example.com" {
		t.Errorf("Server = %q, want %q", loaded.Server, "https://saved.example.com")
	}
	if loaded.APIKey != "saved-key" {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, "saved-key")
	}
}

func TestSaveLocal_CreatesLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg := &Config{
		Org:     "my-org",
		Project: "my-project",
		App:     "my-app",
	}

	err := SaveLocal(cfg)
	if err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}

	// Verify .ancla/config.yaml was created.
	path := filepath.Join(tmpDir, ".ancla", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("local config file not created at %s", path)
	}
}

func TestRemoveLocal(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create local config first.
	cfg := &Config{Org: "test-org"}
	SaveLocal(cfg)

	err := RemoveLocal()
	if err != nil {
		t.Fatalf("RemoveLocal() error: %v", err)
	}

	path := filepath.Join(tmpDir, ".ancla", "config.yaml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("local config file still exists after RemoveLocal()")
	}
}

func TestRemoveLocal_NoLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Should not error when there's nothing to remove.
	err := RemoveLocal()
	if err != nil {
		t.Fatalf("RemoveLocal() with no local config: %v", err)
	}
}

func TestConfig_IsLinked(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"all empty", Config{}, false},
		{"org only", Config{Org: "test"}, true},
		{"project only", Config{Project: "test"}, true},
		{"app only", Config{App: "test"}, true},
		{"all set", Config{Org: "o", Project: "p", App: "a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsLinked(); got != tt.want {
				t.Errorf("IsLinked() = %v, want %v", got, tt.want)
			}
		})
	}
}
