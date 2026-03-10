# now

> Ship with your AI agents — live and in public.

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/opennow-labs/now-cli?color=blue)](https://github.com/opennow-labs/now-cli/releases)
[![License](https://img.shields.io/badge/license-O--Saasy-green)](LICENSE.md)
[![macOS](https://img.shields.io/badge/macOS-supported-black?logo=apple)](https://github.com/opennow-labs/now-cli)
[![Linux](https://img.shields.io/badge/Linux-supported-FCC624?logo=linux&logoColor=black)](https://github.com/opennow-labs/now-cli)
[![Windows](https://img.shields.io/badge/Windows-supported-0078D4?logo=windows&logoColor=white)](https://github.com/opennow-labs/now-cli)

Auto-detects what you're working on and keeps your [opennow.dev](https://opennow.dev) status evergreen — no manual updates needed.

<!-- TODO: screenshot or gif demo -->

## Why

You're shipping fast — solo or with AI agents — but nobody sees the momentum. Your board goes stale, your teammates wonder what you're up to, and updating status manually breaks flow.

`now` runs quietly in the background, detects your active context (editor, music, video), and keeps your presence live. Zero friction, full transparency.

## How It Works

1. **Detect** — reads your active app, music, and video from the OS
2. **Push** — sends a status update to opennow.dev every 30 seconds
3. **Share** — your board stays green and your teammates see what you're building

## Install

```bash
# macOS
brew install opennow-labs/tap/now-cli

# Linux / macOS (script)
curl -fsSL https://opennow.dev/install.sh | sh

# Windows
irm https://opennow.dev/install.ps1 | iex

# From source
go install github.com/opennow-labs/now-cli@latest
```

## Quick Start

```bash
now login    # opens browser for device flow auth
now start    # auto-detect context, push every 30s
```

## Commands

| Command | Description |
|---|---|
| `now login` | Authenticate via device flow |
| `now start` | Start daemon with auto-push |
| `now stop` | Stop the daemon |
| `now status` | Show current board status |
| `now detect` | Print detected context |
| `now push [msg]` | Detect + push (or send a custom message) |
| `now hook` | Manage git hooks |
| `now wrap` | Run a command, push its result |
| `now config` | Open config in your editor |
| `now upgrade` | Self-update to latest release |
| `now uninstall` | Clean removal of now |
| `now version` | Print version info |

## Features

### Context Detection

| Signal | macOS | Linux | Windows |
|---|---|---|---|
| Active app | lsappinfo | xdotool + xprop | PowerShell |
| Window title | osascript | xdotool | PowerShell |
| Music | nowplaying-helper / osascript | playerctl | GlobalSystemMediaTransportControls |
| Video | nowplaying-helper / window title | window title | window title |

Supports Spotify, Apple Music, YouTube, Netflix, Twitch, and many more. Missing signals are silently skipped.

### Git Hooks

Automatically push status on git events:

```bash
now hook install                                    # post-commit hook
now hook install --hooks post-commit,pre-push       # multiple hooks
now hook install --template "Shipped: {commit_msg}" # custom message
```

Hooks are appended (never overwritten) and managed via `# now:start` / `# now:end` markers.

### Command Wrapper

Run any command and push its outcome as status:

```bash
now wrap -- make build                   # "make completed" or "make failed (exit 2)"
now wrap --name "Deploy" -- ./deploy.sh  # "Deploy completed"
```

Template variables: `{cmd}`, `{name}`, `{exit_code}`, `{duration}`. Exit code is preserved.

### System Tray

The daemon shows a system tray icon with current status, now playing info, pause/resume, settings UI, and update notifications.

## Privacy

You stay in full control of what leaves your machine.

| Data | Detected locally | Sent to server | Toggle |
|---|---|---|---|
| Active app name | Yes | Only if `send_app: true` | `send_app` |
| Activity label | Yes | Only if `send_app: true` | `send_app` |
| Window title | Yes | **Never** | — |
| Music artist & track | Yes | Only if `send_music: true` | `send_music` |
| Video content | Yes | Only if `send_watching: true` | `send_watching` |
| OS & architecture | Yes | Only if `telemetry: true` | `telemetry` |

Window titles are never transmitted. All toggles can be set to `false` to share only your green presence. See [Privacy Controls](docs/configuration.md#privacy-controls) for details.

## Configuration

Config lives at `~/.config/now/config.yml`. Quick example:

```yaml
template: "{activity}"
interval: 30s

activity_rules:
  - match: ["Cursor", "Code", "Zed"]
    activity: "Vibe coding"

ignore:
  - "1Password"
```

Full reference with all fields, activity rules, and privacy options: **[docs/configuration.md](docs/configuration.md)**

## Development

```bash
go build -o now .
go test ./...
```

## License

[O-Saasy](LICENSE.md)
