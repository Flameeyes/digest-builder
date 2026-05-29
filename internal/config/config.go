// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

// Package config handles YAML configuration with environment variable
// overrides for digest-builder.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration.
type Config struct {
	Wallabag WallabagConfig `yaml:"wallabag"`
	Digest   DigestConfig   `yaml:"digest"`
	Output   OutputConfig   `yaml:"output"`
	State    StateConfig    `yaml:"state"`
	Proxy    ProxyConfig    `yaml:"proxy"`
}

// WallabagConfig holds Wallabag API connection details.
type WallabagConfig struct {
	URL          string `yaml:"url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
}

// DigestConfig holds article selection and section ordering.
type DigestConfig struct {
	PendingTag      string          `yaml:"pending_tag"`
	MinWordCount    int             `yaml:"min_word_count"`
	Sections        []SectionConfig `yaml:"sections"`
	ExcludePatterns []string        `yaml:"exclude_patterns"`
}

// SectionConfig maps Wallabag tags to digest sections.
type SectionConfig struct {
	Key   string   `yaml:"key"`
	Label string   `yaml:"label"`
	Tags  []string `yaml:"tags"`
}

// OutputConfig controls where EPUBs are written.
type OutputConfig struct {
	Dir      string `yaml:"dir"`
	Filename string `yaml:"filename_format"`
}

// FilenameFormat returns a printf-style pattern with one %s for the date stamp.
func (o OutputConfig) FilenameFormat() string {
	if o.Filename == "" {
		return "digest-%s.epub"
	}
	s := strings.ReplaceAll(o.Filename, "{{.Date}}", "%s")
	if !strings.Contains(s, "%s") {
		s = "digest-%s.epub"
	}
	return s
}

// StateConfig controls OAuth token persistence.
type StateConfig struct {
	File string `yaml:"file"`
}

// ProxyConfig optionally routes HTTP through a proxy.
type ProxyConfig struct {
	HTTP  string `yaml:"http"`
	HTTPS string `yaml:"https"`
}

// Load reads config from a YAML file (if provided) and applies
// environment variable overrides.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	}

	applyEnv(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaults() Config {
	return Config{
		Digest: DigestConfig{
			PendingTag: "digest-pending",
			Sections: []SectionConfig{
				{Key: "ai_case", Label: "AI & CASE Tools", Tags: []string{"AI & CASE Tools"}},
				{Key: "general_tech", Label: "Hacker News & Tech", Tags: []string{"Hacker News & Tech"}},
				{Key: "technical_writing", Label: "Technical Writing", Tags: []string{"Technical Writing"}},
				{Key: "photography", Label: "Photography", Tags: []string{"Photography"}},
				{Key: "ft_subscriber", Label: "Financial Times", Tags: []string{"Financial Times"}},
				{Key: "hbr_subscriber", Label: "Harvard Business Review", Tags: []string{"Harvard Business Review"}},
				{Key: "ftav_further_reading", Label: "FTAV Further Reading", Tags: []string{"FTAV Further Reading"}},
			},
		},
		Output: OutputConfig{
			Dir:      "/output",
			Filename: "digest-{{.Date}}.epub",
		},
		State: StateConfig{
			File: "/state/token.json",
		},
	}
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("WALLABAG_URL"); v != "" {
		cfg.Wallabag.URL = v
	}
	if v := os.Getenv("WALLABAG_CLIENT_ID"); v != "" {
		cfg.Wallabag.ClientID = v
	}
	if v := os.Getenv("WALLABAG_CLIENT_SECRET"); v != "" {
		cfg.Wallabag.ClientSecret = v
	}
	if v := os.Getenv("WALLABAG_USERNAME"); v != "" {
		cfg.Wallabag.Username = v
	}
	if v := os.Getenv("WALLABAG_PASSWORD"); v != "" {
		cfg.Wallabag.Password = v
	}
	if v := os.Getenv("DIGEST_OUTPUT_DIR"); v != "" {
		cfg.Output.Dir = v
	}
	if v := os.Getenv("DIGEST_STATE_FILE"); v != "" {
		cfg.State.File = v
	}
	if v := os.Getenv("HTTP_PROXY"); v != "" {
		cfg.Proxy.HTTP = v
	}
	if v := os.Getenv("HTTPS_PROXY"); v != "" {
		cfg.Proxy.HTTPS = v
	}
}

func validate(cfg *Config) error {
	if cfg.Wallabag.URL == "" {
		return fmt.Errorf("wallabag.url is required (set in config or WALLABAG_URL)")
	}
	if cfg.Wallabag.ClientID == "" {
		return fmt.Errorf("wallabag.client_id is required (set in config or WALLABAG_CLIENT_ID)")
	}
	if cfg.Wallabag.ClientSecret == "" {
		return fmt.Errorf("wallabag.client_secret is required (set in config or WALLABAG_CLIENT_SECRET)")
	}
	if cfg.Wallabag.Username == "" {
		return fmt.Errorf("wallabag.username is required (set in config or WALLABAG_USERNAME)")
	}
	if cfg.Wallabag.Password == "" {
		return fmt.Errorf("wallabag.password is required (set in config or WALLABAG_PASSWORD)")
	}
	return nil
}
