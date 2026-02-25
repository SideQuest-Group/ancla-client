// Command gen-docs generates Starlight-compatible Markdown reference pages
// from the ancla CLI's cobra command tree, organized into subdirectories
// so Starlight auto-generates grouped sidebar navigation.
//
// Usage:
//
//	go run ./cmd/gen-docs --out docs/src/content/docs/cli
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/SideQuest-Group/ancla-client/internal/cli"
)

func main() {
	out := "docs/src/content/docs/cli"
	for i, arg := range os.Args[1:] {
		if arg == "--out" && i+1 < len(os.Args)-1 {
			out = os.Args[i+2]
		}
	}

	rootCmd := cli.RootCmd()
	rootCmd.DisableAutoGenTag = true

	// Pre-compute which commands have subcommands (group parents).
	// These get placed into subdirectories: ancla_apps → cli/apps/
	groups := collectGroups(rootCmd)

	// Generate flat into a temp dir first, then reorganize.
	tmp, err := os.MkdirTemp("", "ancla-docs-*")
	if err != nil {
		log.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, ".md")
		parts := strings.SplitN(base, "_", 3)

		switch {
		case len(parts) >= 3:
			// Subcommand: ancla_apps_deploy → /cli/apps/ancla_apps_deploy/
			return "/cli/" + parts[1] + "/" + base + "/"
		case len(parts) == 2 && groups[parts[1]]:
			// Group parent: ancla_apps → /cli/apps/ancla_apps/
			return "/cli/" + parts[1] + "/" + base + "/"
		default:
			// Root or standalone: ancla, ancla_login → /cli/ancla_login/
			return "/cli/" + base + "/"
		}
	}

	if err := doc.GenMarkdownTreeCustom(rootCmd, tmp, frontmatter, linkHandler); err != nil {
		log.Fatalf("generating docs: %v", err)
	}

	// Clean output directory of old generated content (keep .gitkeep).
	if entries, err := os.ReadDir(out); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				os.RemoveAll(filepath.Join(out, e.Name()))
			} else if strings.HasSuffix(e.Name(), ".md") {
				os.Remove(filepath.Join(out, e.Name()))
			}
		}
	}
	os.MkdirAll(out, 0o755)

	// Reorganize into subdirectories for Starlight grouping.
	entries, _ := os.ReadDir(tmp)
	count := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(e.Name(), ".md")
		parts := strings.SplitN(name, "_", 3)

		destDir := out
		switch {
		case len(parts) >= 3:
			destDir = filepath.Join(out, parts[1])
		case len(parts) == 2 && groups[parts[1]]:
			destDir = filepath.Join(out, parts[1])
		}

		os.MkdirAll(destDir, 0o755)

		src, _ := os.ReadFile(filepath.Join(tmp, e.Name()))
		if err := os.WriteFile(filepath.Join(destDir, e.Name()), src, 0o644); err != nil {
			log.Fatalf("writing %s: %v", e.Name(), err)
		}
		count++
	}

	fmt.Printf("Generated %d CLI reference pages in %s\n", count, out)
}

// collectGroups walks the cobra command tree and returns the set of
// command names (Use field) that have subcommands.
func collectGroups(root *cobra.Command) map[string]bool {
	groups := make(map[string]bool)
	for _, cmd := range root.Commands() {
		if cmd.HasSubCommands() {
			groups[cmd.Name()] = true
		}
	}
	return groups
}

// frontmatter returns Starlight-compatible YAML frontmatter.
// Uses short titles: "ancla_apps_deploy" → "deploy".
func frontmatter(filename string) string {
	name := strings.TrimSuffix(filepath.Base(filename), ".md")
	parts := strings.Split(name, "_")

	var title string
	switch len(parts) {
	case 1:
		title = "Overview"
	case 2:
		title = parts[1]
	default:
		title = strings.Join(parts[2:], " ")
	}

	return fmt.Sprintf(`---
title: "%s"
---

`, title)
}
