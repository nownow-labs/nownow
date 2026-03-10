# Configuration

`now` stores its configuration at `~/.config/now/config.yml` (or `$XDG_CONFIG_HOME/now/config.yml`).

Run `now config` to open the file in your editor.

## Full Example

```yaml
endpoint: https://opennow.dev
token: now_xxx

# Status template — available: {app}, {title}, {music}, {music.artist}, {music.track}, {watching}, {activity}
template: "{activity}"

# Watch interval
interval: 30s

# Activity rules (exact match, case-insensitive)
activity_rules:
  - match: ["Visual Studio Code", "Code", "Cursor", "Windsurf", "Zed"]
    activity: "Vibe coding"
  - match: ["Xcode", "Android Studio"]
    activity: "Building an app"
  - match: ["Terminal", "iTerm2", "Warp", "Alacritty", "kitty"]
    activity: "Hacking away"
  - match: ["Google Chrome", "Safari", "Arc", "Firefox", "Brave Browser"]
    activity: "Down the rabbit hole"
  - match: ["Figma", "Sketch", "Framer"]
    activity: "Pushing pixels"
  - match: ["Slack", "Discord", "Telegram", "WeChat"]
    activity: "In conversation"
  - match: ["Notion", "Obsidian", "Bear", "Notes"]
    activity: "Capturing thoughts"

# Privacy controls (all enabled by default)
telemetry: true       # overall telemetry
send_app: true        # send app name
send_music: true      # send music info
send_watching: true   # send video content

# Automatic update checks
auto_update: true

# Apps to ignore (case-insensitive)
ignore:
  - "1Password"
  - "System Preferences"
  - "System Settings"
```

## Field Reference

### Core

| Field | Type | Default | Description |
|---|---|---|---|
| `endpoint` | string | `https://opennow.dev` | API endpoint. Change for self-hosted instances. |
| `token` | string | — | Auth token, set by `now login`. |
| `template` | string | `{activity}` | Status template. See [Template Variables](#template-variables). |
| `interval` | duration | `30s` | How often the daemon pushes status. |
| `auto_update` | bool | `true` | Check for updates automatically. |

### Template Variables

| Variable | Description |
|---|---|
| `{app}` | Active application name |
| `{title}` | Window title (local only — never sent to server) |
| `{music}` | Full music string (artist – track) |
| `{music.artist}` | Music artist |
| `{music.track}` | Music track name |
| `{watching}` | Video content being watched |
| `{activity}` | Activity label from activity rules |

## Activity Rules

Activity rules map foreground app names to human-readable labels. Rules are matched **top-down**, first match wins.

```yaml
activity_rules:
  - match: ["Visual Studio Code", "Code", "Cursor"]
    activity: "Vibe coding"
  - match: ["Terminal", "iTerm2", "Warp"]
    activity: "Hacking away"
```

- **`match`**: list of app names, exact match, case-insensitive.
- **`activity`**: the label shown in your status.

The default config ships with 40+ rules covering dev tools, browsers, design apps, communication, writing, media, and more. Run `now config` to see and customize them.

## Privacy Controls

Each data type can be independently toggled. When a toggle is off, the corresponding fields are cleared **before** any network request — the data never leaves your machine.

| Field | Default | What it controls |
|---|---|---|
| `telemetry` | `true` | OS & architecture in User-Agent |
| `send_app` | `true` | App name and activity label |
| `send_music` | `true` | Music artist & track |
| `send_watching` | `true` | Video content info |

### Disable everything except presence

```yaml
send_app: false
send_music: false
send_watching: false
telemetry: false
```

This gives you a green dot on the board with zero context shared.

## Ignore List

Block specific apps from being reported entirely. When an ignored app is in the foreground, no status update is sent — your previous status is preserved.

```yaml
ignore:
  - "1Password"
  - "System Preferences"
  - "System Settings"
```

Matching is case-insensitive and supports prefix matching.
