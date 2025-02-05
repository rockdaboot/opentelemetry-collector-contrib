// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otelserializer // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter/internal/serializer/otelserializer"

import (
	"bytes"
	"encoding/json"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter/internal/serializer/otelserializer/serializeprofiles"
)

const (
	allEventsIndex   = "profiling-events-all"
	stackTraceIndex  = "profiling-stacktraces"
	stackFrameIndex  = "profiling-stackframes"
	executablesIndex = "profiling-executables"
)

// SerializeProfile serializes a profile and calls the `pushData` callback for each generated document.
func SerializeProfile(resource pcommon.Resource, scope pcommon.InstrumentationScope, profile pprofile.Profile, pushData func(*bytes.Buffer, string, string) error) error {
	pushDataAsJSON := func(data any, id, index string) error {
		c, err := toJSON(data)
		if err != nil {
			return err
		}
		return pushData(c, id, index)
	}

	data, err := serializeprofiles.Transform(resource, scope, profile)
	if err != nil {
		return err
	}

	for _, payload := range data {
		event := payload.StackTraceEvent

		if event.StackTraceID != "" {
			if err = pushDataAsJSON(event, "", allEventsIndex); err != nil {
				return err
			}
			if err = serializeprofiles.IndexDownsampledEvent(event, pushDataAsJSON); err != nil {
				return err
			}
		}

		if payload.StackTrace.DocID != "" {
			if err = pushDataAsJSON(payload.StackTrace, payload.StackTrace.DocID, stackTraceIndex); err != nil {
				return err
			}
		}

		for _, stackFrame := range payload.StackFrames {
			if err = pushDataAsJSON(stackFrame, stackFrame.DocID, stackFrameIndex); err != nil {
				return err
			}
		}

		for _, executable := range payload.Executables {
			if err = pushDataAsJSON(executable, executable.DocID, executablesIndex); err != nil {
				return err
			}
		}
	}
	return nil
}

func toJSON(d any) (*bytes.Buffer, error) {
	c, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(c), nil
}
