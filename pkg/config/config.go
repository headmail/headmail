// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	SMTP     SMTPConfig     `koanf:"smtp"`
	IMAP     IMAPConfig     `koanf:"imap"`
	Tracking TrackingConfig `koanf:"tracking"`
	Database DatabaseConfig `koanf:"database"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Public struct {
		Port int    `koanf:"port"`
		URL  string `koanf:"url"`
	} `koanf:"public"`
	Admin struct {
		Addr int `koanf:"addr"`
		Port int `koanf:"port"`
	} `koanf:"admin"`
}

// SMTPConfig holds SMTP-related configuration.
type SMTPConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	From     struct {
		Name  string `koanf:"name"`
		Email string `koanf:"email"`
	} `koanf:"from"`
	Send struct {
		BatchSize int `koanf:"batch_size"`
		Throttle  int `koanf:"throttle"`
		Attempts  int `koanf:"attempts"`
	} `koanf:"send"`
}

type IMAPConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	TLS      bool   `koanf:"tls"`
	Mailbox  string `koanf:"mailbox"` // e.g. "INBOX"
}

// TrackingConfig holds tracking-related configuration.
type TrackingConfig struct {
	// ImagePath is an optional path or URL to a tracking image to return for opens.
	// If empty, a built-in 1x1 transparent PNG will be returned.
	ImagePath string `koanf:"image_path"`
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	Type string `koanf:"type"`
	URL  string `koanf:"url"`
}

// Option defines a function that configures a koanf instance.
type Option func(k *koanf.Koanf)

// WithFile creates an option that loads configuration from a file.
func WithFile(path string) Option {
	return func(k *koanf.Koanf) {
		var parser koanf.Parser
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			parser = yaml.Parser()
		} else if strings.HasSuffix(path, ".json") {
			parser = json.Parser()
		} else if strings.HasSuffix(path, ".toml") {
			parser = toml.Parser()
		} else {
			log.Printf("Warning: unsupported config file format for %s", path)
			return
		}

		if err := k.Load(file.Provider(path), parser); err != nil {
			log.Printf("Warning: could not load config file %s: %v", path, err)
		}
	}
}

var envMappings = map[string]string{
	"SMTP_SEND_BATCH_SIZE": "smtp.send.batch_size",
}

// Load loads the configuration using the provided options.
func Load(opts ...Option) (*Config, error) {
	k := koanf.New(".")

	// Set default values
	k.Set("server.public.port", 8080)
	k.Set("server.admin.port", 8081)
	k.Set("database.type", "sqlite")
	k.Set("database.url", "file:data.db?cache=shared&mode=rwc")

	// Apply all options
	for _, opt := range opts {
		opt(k)
	}

	// Load environment variables
	if err := k.Load(env.Provider("HEADMAIL_", ".", func(s string) string {
		envKey := strings.TrimPrefix(s, "HEADMAIL_")
		mapped, ok := envMappings[envKey]
		if ok {
			return mapped
		}
		return strings.Replace(strings.ToLower(envKey), "_", ".", -1)
	}), nil); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
