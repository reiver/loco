# Loco — Design Document

## Overview

**loco** is a command-line tool written in the Go programming-language (golang) that enables users to install and run Fediverse and other self-hosted software as Docker containers.
App recipes are distributed as git repositories hosted on any forge (GitHub, GitLab, Codeberg, self-hosted Forgejo, etc.).

The tool aims to provide a turnkey experience for newcomers while giving power users full control.

---

## Core Concepts

### App Repositories

Each app is defined by a git repository containing:

* **`manifest.plan`** — App metadata, environment variables, data volumes, host config, and install steps (colon-separated INI format with `.plan` extension).
* **`docker-compose.yml`** — Docker Compose file defining the app's containers. All apps use Compose (no raw Dockerfile support).
* **`.env.tmpl`** — Environment variable template. Repos must NOT commit a `.env` file; only the template.
* **Scripts** — Install/setup scripts that run **inside** containers only. Host-side effects are declared in the manifest, not scripted.

### Loco Root

Running `loco init` in any directory creates a `.loco/` directory, making that location a loco root.
Multiple roots can exist on the same machine (e.g., production and testing).

All `loco` commands search parent directories for `.loco/`, similar to how `git` finds `.git/`.

---

## Architecture — Dispatcher Pattern

The `loco` binary is a thin dispatcher, following the same pattern as `git`.
It does not implement sub-commands directly.
Instead:

1. The user runs `loco <subcommand> [args...]`
2. `loco` searches `$PATH` for a binary named `loco-<subcommand>`
3. `loco` executes `loco-<subcommand> [args...]`, passing all remaining arguments through

### Examples

```bash
loco init                          # executes: loco-init
loco install repo@v1.0             # executes: loco-install repo@v1.0
loco apple --banana cherry         # executes: loco-apple --banana cherry
```

### Built-in Sub-Commands

The following `loco-*` binaries ship with loco:

| Binary            | Command              |
|-------------------|----------------------|
| `loco-init`       | `loco init`          |
| `loco-install`    | `loco install`       |
| `loco-update`     | `loco update`        |
| `loco-start`      | `loco start`         |
| `loco-stop`       | `loco stop`          |
| `loco-remove`     | `loco remove`        |
| `loco-list`       | `loco list`          |
| `loco-status`     | `loco status`        |
| `loco-logs`       | `loco logs`          |
| `loco-snapshot`   | `loco snapshot`      |
| `loco-snapshots`  | `loco snapshots`     |
| `loco-restore`    | `loco restore`       |
| `loco-scheduler`  | `loco scheduler`     |

### Third-party extensions

Anyone can create a `loco-*` binary and place it anywhere in `$PATH` to extend loco.
For example, if someone creates `loco-party` and installs it to `/usr/local/bin/`, then `loco party` automatically works.

### Error handling

If `loco-<subcommand>` is not found in `$PATH`:

```
loco: 'party' is not a loco command. See 'loco help'.
```

### Help / discovery

`loco help` (or bare `loco` with no arguments) scans `$PATH` for all `loco-*` binaries and lists available commands — both built-in and third-party.

---

## Directory Structure

```
myserver/
├── .loco/
│   ├── config.plan                                         # Root-level config (INI format, .plan extension)
│   ├── snapshots/                                          # Default snapshot storage
│   ├── apps/
│   │   ├── github.com/reiver/locoverse-nextcloud/          # Cloned app repo
│   │   ├── codeberg.org/greatape/locoverse-greatape/       # Cloned app repo
│   │   └── ...
│   └── data/
│       ├── github.com/reiver/locoverse-nextcloud/
│       │   ├── .env                                        # Generated env file
│       │   ├── nextcloud-files/                            # Persistent data (bind-mounted into container)
│       │   └── nextcloud-db/                               # Persistent data (bind-mounted into container)
│       ├── codeberg.org/greatape/locoverse-greatape/
│       │   ├── .env
│       │   └── ...
│       └── ...
```

### Separation of apps and data

* **`apps/`** contains only cloned git repositories. It is purely git-managed and stays clean.
* **`data/`** contains all user/runtime state: generated `.env` files, persistent data volumes, etc. Data survives container destruction. Backing up `data/` captures the full state.

Both trees use the full repo URL as the directory path (e.g., `github.com/reiver/locoverse-nextcloud`) to ensure uniqueness across forges.

---

## Manifest Format — `manifest.plan`

Colon-separated INI format with `.plan` extension.

```ini
[app]
name: nextcloud
description: Self-hosted cloud storage
version: 1.0.0

[env]
NEXTCLOUD_HOSTNAME: nextcloud-locoverse
NEXTCLOUD_PORT: 20035

[data]
nextcloud-files: /srv/nextcloud/data
nextcloud-db: /var/lib/mysql

[host]
hostname: nextcloud-locoverse
port: 20035
mdns: true

[install]
step1: /scripts/setup.sh
step2: /scripts/filescan-cron.sh
```

### Sections

| Section     | Purpose |
|-------------|---------|
| `[app]`     | App name, description, version. |
| `[env]`     | Environment variable defaults. Used alongside `.env.tmpl` for generating the `.env` file. |
| `[data]`    | Named data volumes. Keys are descriptive names (used as host-side directory names under `data/`). Values are container-side mount points. |
| `[host]`    | Declarative host-side configuration (hostname, default port, mDNS). The CLI interprets these — no host-side scripts. |
| `[install]` | Ordered install steps (numbered keys). Scripts run **inside** the container only. |

---

## Environment Configuration — `.env.tmpl`

App repos commit an `.env.tmpl` file (never `.env`) with `{{PLACEHOLDER}}` syntax for values the CLI resolves at install time:

```
NEXTCLOUD_HOSTNAME={{HOSTNAME}}
NEXTCLOUD_PORT={{PORT}}
NEXTCLOUD_DATA_DIR={{DATA_DIR}}
DB_PASSWORD={{DB_PASSWORD}}
```

The CLI:
1. Reads `.env.tmpl` from the cloned repo in `apps/`.
2. Resolves built-in placeholders (e.g., `{{DATA_DIR}}` → actual path under `.loco/data/`).
3. Prompts the user interactively for remaining values (default behavior), or the user can edit the generated file manually.
4. Writes the generated `.env` to the `data/` side.

Docker Compose is invoked with `--env-file` pointing to the `data/` location.

---

## Root Config — `.loco/config.plan`

Colon-separated INI format with `.plan` extension.
Single source of truth for all loco configuration — both global settings and per-app overrides.

### Per-App Overrides

Any section can have per-app overrides using a git-style quoted section syntax:

```ini
[section]
key: global-default

[section "github.com/reiver/locoverse-nextcloud"]
key: override-for-this-app
```

Per-app sections inherit all values from the global section and only need to specify the values they want to override.

### Full Example

```ini
[network]
hostname: locoverse.local

[auth]
github.com: ghp_xxxxxxxxxxxx
codeberg.org: tok_xxxxxxxxxxxx

[scheduler]
backend: cron

[snapshots]
keep: 5
max-age: 90d
auto: daily
exclude: github.com/someone/app-a, github.com/someone/app-b

[snapshots "github.com/reiver/locoverse-nextcloud"]
keep: 10
auto: hourly
```

### Sections

| Section       | Purpose |
|---------------|---------|
| `[network]`   | Default hostname for mDNS (default: `locoverse.local`). Set during `loco init`. |
| `[auth]`      | Forge authentication tokens for private repositories. Keys are forge hostnames, values are tokens. |
| `[scheduler]` | Scheduling backend for auto-snapshots and other scheduled tasks. See [Scheduler](#scheduler). |
| `[snapshots]` | Snapshot retention policy and auto-snapshot settings. See [Snapshots](#snapshots). |

---

## Networking

### mDNS / Local Access

A single `.local` mDNS hostname is used for all apps, with different ports per app:

* `locoverse.local:20035` → NextCloud
* `locoverse.local:20036` → PeerTube
* `locoverse.local:20037` → Pixelfed

The hostname defaults to `locoverse.local` and is configurable during `loco init`.

**Note**: mDNS does not reliably support subdomains across platforms (Windows cannot resolve `sub.hostname.local`), so flat hostnames are used.

### Inter-App Communication

A shared Docker network (e.g., `locoverse-net`) allows installed apps to communicate with each other by service name.

### Internet Access

A separate tool (outside the scope of loco) will be created later to expose apps to the internet.

### Port Management

Each app specifies a default port in its manifest. The CLI:

* Checks if the default port is available.
* If not (or if the user wants a different port), prompts for an alternative or auto-assigns.
* Stores the assigned port in the generated `.env`.

---

## Commands

### `loco init`

Initializes a loco root in the current directory.
Creates `.loco/` with `config.plan`, `apps/`, and `data/` directories.
Prompts for hostname (default: `locoverse.local`).

### `loco install <repo>[@<version>]`

Full install flow:
1. Resolve the repo URL.
2. Clone the repo to `.loco/apps/<repo-path>/`.
3. Read `manifest.plan` and `.env.tmpl`.
4. Create data directories under `.loco/data/<repo-path>/`.
5. Generate `.env` — prompt interactively for values (default) or allow manual editing.
6. Run `docker compose up -d` with `--env-file` pointing to `data/`.
7. Run install scripts inside the container (in manifest order).
8. Configure mDNS hostname if enabled.

Version pinning with `@` syntax:
```bash
loco install github.com/reiver/locoverse-nextcloud              # default branch
loco install github.com/reiver/locoverse-nextcloud@v1.2.0       # tag
loco install github.com/reiver/locoverse-nextcloud@main         # branch
loco install github.com/reiver/locoverse-nextcloud@abc123def    # commit hash
```

### `loco update <app>`

1. Snapshot data automatically (safety net).
2. `git pull` the app repo.
3. Re-render `.env.tmpl`, preserving existing user values.
4. Rebuild/re-pull Docker image.
5. Restart containers.

### `loco start <app>`

Start a stopped app's containers.

### `loco stop <app>`

Stop a running app's containers.

### `loco remove <app>`

Stop containers and remove them.
Prompt the user about whether to also delete data.

### `loco list`

Show all installed apps with their status (running/stopped) and URL (`locoverse.local:<port>`).

### `loco status <app>`

Detailed status of a specific app: container state, port, data directory size, version/branch, etc.

### `loco logs <app>`

Wraps `docker compose logs` for the app's containers.

### `loco snapshot [<app>] [--output <path>]`

Create a snapshot (archive) of an app's data directory.
If no app is specified, snapshot all apps (excluding any listed in the `exclude` setting).
Snapshots are stored in `.loco/snapshots/` by default. Optionally specify an output directory with `--output`.

### `loco snapshot --delete <snapshot-id>`

Delete a specific snapshot.
Prompts for confirmation unless `--force` or `-y` is passed.

Additional deletion modes:
```bash
loco snapshot --delete --all                      # delete all snapshots
loco snapshot --delete --all nextcloud            # delete all for an app
loco snapshot --delete --older-than 30d           # delete old ones across all apps
loco snapshot --delete --older-than 30d nextcloud # delete old ones for an app
```

### `loco snapshots [<app>]`

List available snapshots.
If an app is specified, list only snapshots for that app.
Shows snapshot name, type (manual/auto/update), date, and size.

### `loco restore <app> <snapshot>`

Restore an app's data from a snapshot:
1. Check that the app is installed (warn if not — suggest `loco install` first).
2. Stop the app's containers.
3. Extract the snapshot into the `data/` directory.
4. Restart containers.

---

## App Reference / Shorthand

Users can reference apps by:

1. **Full repo path** (always works): `github.com/reiver/locoverse-nextcloud`
2. **Last URL segment** (if unambiguous): `locoverse-nextcloud`
3. **Manifest app name** (if unambiguous): `nextcloud`

Resolution order: exact full path → last segment → manifest name.
If a shorthand matches multiple installed apps, the CLI errors with a list of matches and asks the user to be more specific.

---

## Install Script Security Boundary

* **Scripts run inside containers only.** No scripts execute on the host.
* **Host-side effects are declarative.** The `[host]` section of `manifest.plan` declares what the CLI should configure on the host (mDNS, etc.). The CLI interprets and applies these — no arbitrary host execution.

---

## Snapshots

### Snapshot Naming

Snapshots are named with the app name, timestamp, and type indicator:

```
nextcloud-20260324-1430-manual.tar.gz       # created by user via loco snapshot
nextcloud-20260324-0300-auto.tar.gz         # created by auto-snapshot schedule
nextcloud-20260323-1200-update.tar.gz       # created automatically before loco update
```

### Snapshot Configuration

Configured in `.loco/config.plan` under `[snapshots]`:

```ini
[snapshots]
keep: 5
max-age: 90d
auto: daily
exclude: github.com/someone/app-a, github.com/someone/app-b
```

| Key       | Purpose |
|-----------|---------|
| `keep`    | Minimum number of snapshots to retain per app, regardless of age. |
| `max-age` | Prune snapshots older than this — but only if more than `keep` snapshots remain. `keep` always wins. |
| `auto`    | Auto-snapshot frequency (`hourly`, `daily`, `weekly`). Requires a scheduler backend to be configured. |
| `exclude` | Comma-separated list of app repo paths to exclude from global auto-snapshots. |

### Retention Logic

The `keep` setting takes priority over `max-age` to ensure users are never left with zero snapshots:

* Always retain at least `keep` snapshots per app, regardless of their age.
* Beyond `keep`, prune any snapshot older than `max-age`.
* Example: `keep: 5`, `max-age: 30d`, and 5 snapshots all 60 days old → all 5 are kept.
* Example: `keep: 5`, `max-age: 30d`, and 8 snapshots with 3 older than 30 days → the 3 old ones are pruned (5 remain).

Retention is applied automatically when new snapshots are created (both manual and auto).

### Per-app Overrides

Per-app snapshot settings override globals using the quoted section syntax:

```ini
[snapshots "github.com/reiver/locoverse-nextcloud"]
keep: 10
auto: hourly
```

This app keeps 10 snapshots and runs hourly auto-snapshots.
It inherits `max-age: 90d` from the global `[snapshots]` section.

---

## Scheduler

The scheduler is a pluggable subsystem for auto-snapshots and other scheduled tasks.
The `loco` tool defines *what* to run and *when*; the scheduler backend handles *how*.

### Configuration

```ini
[scheduler]
backend: cron
```

### Backends

| Backend          | Description |
|------------------|-------------|
| `cron`           | Uses system cron jobs. |
| `systemd-timer`  | Uses systemd timer units. |
| `temporal`       | Uses a Temporal server for scheduling. Requires `endpoint` setting. |

Custom backends can be added via third-party `loco-scheduler-<backend>` binaries.

Example with Temporal:
```ini
[scheduler]
backend: temporal
endpoint: temporal.example.com:7233
```

The scheduler manages tasks defined by other settings (e.g., the `auto` key in `[snapshots]`).

---

## Private Repository Support

Forge authentication tokens are stored in `.loco/config.plan` under `[auth]`.
The CLI uses these tokens when cloning private repositories.
Supported from day one.

---

## Technology

* **Language**: Go
* **Container runtime**: Docker with Docker Compose (required)
* **Config format**: Colon-separated INI with `.plan` extension
* **Templating**: `{{PLACEHOLDER}}` syntax in `.env.tmpl`
* **Local networking**: Avahi/mDNS for `.local` hostname resolution
* **Git**: Used for cloning and updating app repositories from any forge
