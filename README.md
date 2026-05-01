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

[Watchtower](https://github.com/containrrr/watchtower) was the go-to tool for automatically updating Docker containers. It was archived in late 2024 and is no longer maintained. As Docker Engine evolves, Watchtower will break due to API incompatibilities.

**Vigil** picks up where Watchtower left off:

- Updated to Docker SDK v27 (fixes "client version too old" errors)
- Go 1.25
- Negotiates Docker API version with the daemon (no more pinned version)
- Multi-arch images (amd64 + arm64), built natively
- Rolling restart is the default (limits blast radius to one container)
- Automatic rollback to the previous image if a recreated container fails to start
- TLS verification enforced for registry digest checks
- New `dev.vigil.*` labels and `VIGIL_*` env vars
- Fully backward-compatible with existing Watchtower labels and env vars

## Quick Start

Vigil is a drop-in replacement for Watchtower:

```bash
docker run --detach \
    --name vigil \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    ghcr.io/nitroxaddict/vigil
```

## Migration from Watchtower

1. Stop your Watchtower container
2. Replace the image with `ghcr.io/nitroxaddict/vigil`
3. Start the container

That's it. All your existing `com.centurylinklabs.watchtower.*` labels and `WATCHTOWER_*` env vars continue to work unchanged. You can migrate to Vigil-branded names at your own pace.

## Configuration

`VIGIL_*` env vars and `dev.vigil.*` labels are the canonical names. The original `WATCHTOWER_*` env vars and `com.centurylinklabs.watchtower.*` labels are accepted for backward compatibility but **will be removed in a future release** — prefer the Vigil names. When both are set, the Vigil name wins.

This applies to **every** env var and label, including ones not enumerated in the tables below.

### Environment Variables

| Vigil | Watchtower (legacy) | Description |
|-------|---------------------|-------------|
| `VIGIL_POLL_INTERVAL` | `WATCHTOWER_POLL_INTERVAL` | Poll interval in seconds (default: 86400) |
| `VIGIL_SCHEDULE` | `WATCHTOWER_SCHEDULE` | Cron expression for update schedule |
| `VIGIL_CLEANUP` | `WATCHTOWER_CLEANUP` | Remove old images after updating |
| `VIGIL_LABEL_ENABLE` | `WATCHTOWER_LABEL_ENABLE` | Only update containers with enable label |
| `VIGIL_MONITOR_ONLY` | `WATCHTOWER_MONITOR_ONLY` | Only monitor, don't update |
| `VIGIL_BATCH_RESTART` | `WATCHTOWER_BATCH_RESTART` | Stop all containers before restarting (default: rolling) |
| `VIGIL_HTTP_API_UPDATE` | `WATCHTOWER_HTTP_API_UPDATE` | Enable HTTP API for on-demand updates |
| `VIGIL_NO_PULL` | `WATCHTOWER_NO_PULL` | Do not pull new images |
| `VIGIL_NO_RESTART` | `WATCHTOWER_NO_RESTART` | Do not restart containers |
| `VIGIL_INCLUDE_STOPPED` | `WATCHTOWER_INCLUDE_STOPPED` | Include stopped containers |
| `VIGIL_LIFECYCLE_HOOKS` | `WATCHTOWER_LIFECYCLE_HOOKS` | Enable pre/post update hooks |

### Notification Environment Variables

| Vigil | Watchtower (legacy) | Description |
|-------|---------------------|-------------|
| `VIGIL_NOTIFICATIONS` | `WATCHTOWER_NOTIFICATIONS` | Notification types to enable (`email`, `slack`, `gotify`, `msteams`, `shoutrrr`) |
| `VIGIL_NOTIFICATIONS_LEVEL` | `WATCHTOWER_NOTIFICATIONS_LEVEL` | Log level for notifications (default: `info`) |
| `VIGIL_NOTIFICATIONS_HOSTNAME` | `WATCHTOWER_NOTIFICATIONS_HOSTNAME` | Hostname shown in notification title |
| `VIGIL_NOTIFICATION_REPORT` | `WATCHTOWER_NOTIFICATION_REPORT` | Use the structured per-container report template instead of the log-dump default |
| `VIGIL_NOTIFICATION_TEMPLATE` | `WATCHTOWER_NOTIFICATION_TEMPLATE` | Custom Go template for the notification body |
| `VIGIL_NOTIFICATION_URL` | `WATCHTOWER_NOTIFICATION_URL` | Shoutrrr notification URL(s) |
| `VIGIL_NOTIFICATION_TITLE_TAG` | `WATCHTOWER_NOTIFICATION_TITLE_TAG` | Prefix added to notification subject/title |

Email-, Slack-, MS Teams- and Gotify-specific options (SMTP server, hook URLs, tokens, etc.) follow the same pattern: every `WATCHTOWER_NOTIFICATION_*` variable accepts a `VIGIL_NOTIFICATION_*` equivalent. See the [original Watchtower notification docs](https://containrrr.dev/watchtower/notifications/) for the full reference.

### Container Labels

| Vigil | Watchtower (legacy) | Description |
|-------|---------------------|-------------|
| `dev.vigil.enable` | `com.centurylinklabs.watchtower.enable` | Enable/disable monitoring |
| `dev.vigil.monitor-only` | `com.centurylinklabs.watchtower.monitor-only` | Monitor only, don't update |
| `dev.vigil.no-pull` | `com.centurylinklabs.watchtower.no-pull` | Skip pulling new images |
| `dev.vigil.stop-signal` | `com.centurylinklabs.watchtower.stop-signal` | Custom stop signal |
| `dev.vigil.depends-on` | `com.centurylinklabs.watchtower.depends-on` | Container dependencies |
| `dev.vigil.scope` | `com.centurylinklabs.watchtower.scope` | Monitoring scope |

See the [original Watchtower documentation](https://containrrr.dev/watchtower) for full details on all options.

## Original Contributors

Vigil is built on the excellent work of the [Watchtower contributors](https://github.com/containrrr/watchtower#contributors). This project would not exist without them.

## License

Apache-2.0 -- same as the original Watchtower project.
