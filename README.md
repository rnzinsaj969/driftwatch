# driftwatch

A daemon that monitors infrastructure config files for unexpected drift and sends alerts via webhook.

---

## Installation

```bash
go install github.com/yourorg/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a config file (`driftwatch.yaml`) defining the files to watch and your webhook endpoint:

```yaml
interval: 60s
webhook: "https://hooks.example.com/alerts"
paths:
  - /etc/nginx/nginx.conf
  - /etc/app/config.toml
  - /opt/service/settings.yaml
```

Then start the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When a monitored file changes unexpectedly, driftwatch sends a POST request to your webhook with details about the drift, including the file path, timestamp, and a diff summary.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `driftwatch.yaml` | Path to config file |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`) |
| `--dry-run` | `false` | Detect drift without sending alerts |

---

## How It Works

1. On startup, driftwatch hashes each watched file to establish a baseline.
2. At the configured interval, it re-hashes each file and compares against the baseline.
3. If drift is detected, an alert payload is sent to the configured webhook.

---

## License

MIT © yourorg