// Package config handles CLI configuration stored at ~/.ancla/config.yaml
// with optional per-directory overrides from .ancla/config.yaml in the
// current directory or any parent.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the CLI configuration.
type Config struct {
	Server   string `mapstructure:"server"`
	APIKey   string `mapstructure:"api_key"`
	Username string `mapstructure:"username"`
	Email    string `mapstructure:"email"`

	// Link context — stored in local .ancla/config.yaml only
	Org     string `mapstructure:"org"`
	Project string `mapstructure:"project"`
	App     string `mapstructure:"app"`
}

// homeConfigDir returns the path to ~/.ancla/.
func homeConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ancla"
	}
	return filepath.Join(home, ".ancla")
}

// findLocalConfigDir walks from cwd upward looking for a .ancla/ directory.
// Returns the path if found, or empty string if none exists.
func findLocalConfigDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ".ancla")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Load reads configuration with the following precedence (highest first):
//  1. CLI flags (--server, --api-key)
//  2. Environment variables (ANCLA_SERVER, ANCLA_API_KEY)
//  3. Local .ancla/config.yaml (nearest parent directory)
//  4. ~/.ancla/config.yaml
//  5. Built-in defaults
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("ANCLA")
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("server", "https://ancla.dev")
	v.SetDefault("api_key", "")

	// Load global config first (~/.ancla/config.yaml)
	v.AddConfigPath(homeConfigDir())
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	// Layer local config on top (.ancla/config.yaml from cwd or parent)
	if localDir := findLocalConfigDir(); localDir != "" {
		local := viper.New()
		local.SetConfigName("config")
		local.SetConfigType("yaml")
		local.AddConfigPath(localDir)
		if err := local.ReadInConfig(); err == nil {
			if err := v.MergeConfigMap(local.AllSettings()); err != nil {
				return nil, fmt.Errorf("merging local config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// FilePath returns the active config file path. If a local .ancla/ exists
// in cwd or a parent, returns that; otherwise returns ~/.ancla/config.yaml.
func FilePath() string {
	if localDir := findLocalConfigDir(); localDir != "" {
		return filepath.Join(localDir, "config.yaml")
	}
	return filepath.Join(homeConfigDir(), "config.yaml")
}

// Paths returns the global and local config file paths.
// Local path is empty if no .ancla/ directory was found in cwd or parents.
func Paths() (global string, local string) {
	global = filepath.Join(homeConfigDir(), "config.yaml")
	if localDir := findLocalConfigDir(); localDir != "" {
		local = filepath.Join(localDir, "config.yaml")
	}
	return
}

// Save writes the current configuration to ~/.ancla/config.yaml.
func Save(cfg *Config) error {
	dir := homeConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	v := viper.New()
	v.Set("server", cfg.Server)
	v.Set("api_key", cfg.APIKey)
	if cfg.Username != "" {
		v.Set("username", cfg.Username)
	}
	if cfg.Email != "" {
		v.Set("email", cfg.Email)
	}
	path := filepath.Join(dir, "config.yaml")
	return v.WriteConfigAs(path)
}

// SaveLocal writes link context (org, project, app) to .ancla/config.yaml
// in the current working directory, creating the directory if needed.
func SaveLocal(cfg *Config) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	localDir := filepath.Join(dir, ".ancla")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return fmt.Errorf("creating .ancla directory: %w", err)
	}
	v := viper.New()
	if cfg.Org != "" {
		v.Set("org", cfg.Org)
	}
	if cfg.Project != "" {
		v.Set("project", cfg.Project)
	}
	if cfg.App != "" {
		v.Set("app", cfg.App)
	}
	path := filepath.Join(localDir, "config.yaml")
	return v.WriteConfigAs(path)
}

// RemoveLocal deletes the .ancla/config.yaml in the current working directory.
func RemoveLocal() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	path := filepath.Join(dir, ".ancla", "config.yaml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing local config: %w", err)
	}
	// Try to remove .ancla dir if empty
	os.Remove(filepath.Join(dir, ".ancla"))
	return nil
}

// IsLinked returns true if the config has any link context set.
func (c *Config) IsLinked() bool {
	return c.Org != "" || c.Project != "" || c.App != ""
}
