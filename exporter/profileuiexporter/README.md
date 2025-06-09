# Profile UI Exporter

| Status                   |                   |
| ------------------------ |-------------------|
| Stability                | [development]     |
| Distributions            | [contrib]         |
| Issues                   | [Issues]          |

**TODO: This is a new exporter and this README is a work in progress.**

This exporter receives pprof profiles and hosts a simple web UI to display them.

## Configuration

The following configuration options are available:

- `http_port` (default: `8080`): The port on which the UI server will listen.

Example:

```yaml
exporters:
  profileui:
    http_port: 8081
```

[development]: https://github.com/open-telemetry/opentelemetry-collector#development
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[Issues]: https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aissue+is%3Aopen+label%3Aexporter%2Fprofileui
