// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package profileuiexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/profileuiexporter"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/profileuiexporter/internal/metadata"
)

// NewFactory creates a factory for the Profile UI exporter.
func NewFactory() exporter.Factory {
	return xexporter.NewFactory(
		component.MustNewType(metadata.Type),
		createDefaultConfig,
		xexporter.WithProfiles(createProfilesExporter, metadata.ProfilesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		HTTPPort: 8080, // Default port for the UI
	}
}

func createProfilesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (xexporter.Profiles, error) {
	pec := cfg.(*Config)
	pe, err := newProfilesExporter(pec, set)
	if err != nil {
		return nil, err
	}

	return xexporterhelper.NewProfilesExporter(
		ctx,
		set,
		cfg,
		pe.consumeProfiles,
		exporterhelper.WithStart(pe.Start),
		exporterhelper.WithShutdown(pe.Shutdown),
		// TODO: Check if MutatesData should be true or false
		// For now, assuming it doesn't mutate data, similar to fileexporter.
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
	)
}
