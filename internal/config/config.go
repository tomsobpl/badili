// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type ExporterConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Type    string `mapstructure:"type"`
}

type ListenerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Type    string `mapstructure:"type"`
}

type PackagerConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type Config struct {
	Exporter ExporterConfig `mapstructure:"exporter"`
	Listener ListenerConfig `mapstructure:"listener"`
	Packager PackagerConfig `mapstructure:"packager"`
}

func load(ctx context.Context) (*Config, error) {
	var cfg Config
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("BADILI")
	v.AutomaticEnv()

	v.SetDefault("exporter.enabled", ExporterEnabledDefault)
	v.SetDefault("exporter.port", ExporterPortDefault)
	v.SetDefault("exporter.type", ExporterTypeDefault)

	v.SetDefault("listener.enabled", ListenerEnabledDefault)
	v.SetDefault("listener.port", ListenerPortDefault)
	v.SetDefault("listener.type", ListenerTypeDefault)

	v.SetDefault("packager.enabled", PackagerEnabledDefault)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.InfoContext(ctx, "config file not found, proceeding with defaults", "err", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}

func InitConfiguration(ctx context.Context) (*Config, error) {
	cfg, err := load(ctx)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
