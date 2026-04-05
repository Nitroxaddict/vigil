<div align="center">

  # Vigil

  A community-maintained fork of [Watchtower](https://github.com/containrrr/watchtower).
  <br/>
  Automatically update running Docker container images.
  <br/><br/>

  [![CI](https://github.com/Nitroxaddict/vigil/actions/workflows/pull-request.yml/badge.svg)](https://github.com/Nitroxaddict/vigil/actions/workflows/pull-request.yml)
  [![Apache-2.0 License](https://img.shields.io/github/license/Nitroxaddict/vigil.svg)](https://www.apache.org/licenses/LICENSE-2.0)

</div>

## Why Vigil?

[Watchtower](https://github.com/containrrr/watchtower) was the go-to tool for automatically updating Docker containers. It was archived in late 2024 and is no longer maintained. As Docker Engine evolves, Watchtower will eventually break due to API incompatibilities.

**Vigil** picks up where Watchtower left off:

- Updated to Docker SDK v27 (fixes "client version too old" errors)
- Go 1.22+
- Docker API version bumped from 1.25 to 1.44
- Fully backward-compatible with existing Watchtower configurations and labels

## Quick Start

Vigil is a drop-in replacement for Watchtower. It uses the same labels, environment variables, and configuration:

```bash
docker run --detach \
    --name vigil \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    ghcr.io/nitroxaddict/vigil
```

All existing `com.centurylinklabs.watchtower.*` labels continue to work.

## Migration from Watchtower

1. Stop your Watchtower container
2. Replace the image with `ghcr.io/nitroxaddict/vigil`
3. Start the container -- all your labels and config carry over

## Configuration

Vigil supports the same configuration options as Watchtower. See the [original Watchtower documentation](https://containrrr.dev/watchtower) for details.

Key environment variables:

| Variable | Description |
|----------|-------------|
| `WATCHTOWER_POLL_INTERVAL` | Poll interval in seconds (default: 86400) |
| `WATCHTOWER_SCHEDULE` | Cron expression for update schedule |
| `WATCHTOWER_CLEANUP` | Remove old images after updating |
| `WATCHTOWER_LABEL_ENABLE` | Only update containers with enable label |
| `WATCHTOWER_MONITOR_ONLY` | Only monitor, don't update |
| `WATCHTOWER_ROLLING_RESTART` | Restart containers one at a time |
| `WATCHTOWER_HTTP_API_UPDATE` | Enable HTTP API for on-demand updates |

## Original Contributors

Vigil is built on the excellent work of the [Watchtower contributors](https://github.com/containrrr/watchtower#contributors). This project would not exist without them.

## License

Apache-2.0 -- same as the original Watchtower project.
