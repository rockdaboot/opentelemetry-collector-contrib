// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package profileuiexporter

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for the Profile UI exporter.
type Config struct {
	// HTTPPort is the port on which the UI server will listen.
	HTTPPort int `mapstructure:"http_port"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid.
func (cfg *Config) Validate() error {
	if cfg.HTTPPort <= 0 || cfg.HTTPPort > 65535 {
		return errors.New("http_port must be a valid port number (1-65535)")
	}
	return nil
}
