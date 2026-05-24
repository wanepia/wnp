# wnp — Wanepia CLI

Command-line interface for [Wanepia](https://wanepia.com) — manage blueprints, entities, uptime checks, alerts, and your team from the terminal.

---

## Install

**Homebrew (macOS / Linux)**

```bash
brew install wanepia/tap/wnp
```

**Download binary** — grab the latest release from [GitHub Releases](https://github.com/wanepia/wnp/releases):

| Platform | Archive |
|----------|---------|
| macOS (arm64) | `wnp_darwin_arm64.tar.gz` |
| macOS (amd64) | `wnp_darwin_amd64.tar.gz` |
| Linux (amd64) | `wnp_linux_amd64.tar.gz` |
| Windows (amd64) | `wnp_windows_amd64.zip` |

**Build from source**

```bash
git clone https://github.com/wanepia/wnp
cd wnp
go install .
```

---

## Getting started

```bash
# Point at your Wanepia instance and log in
wnp config set-url https://api.wanepia.com
wnp login you@example.com

# Or set the token directly (CI / scripts)
wnp config set-token <api-key>

# Check that everything works
wnp status show
```

Config is saved to `~/.config/wnp/config.yaml`. All flags can override it per-command:

```bash
wnp --url https://staging.wanepia.com --token abc123 blueprints list
```

---

## Command reference

### Global flags

| Flag | Description |
|------|-------------|
| `--url <url>` | API base URL (overrides config) |
| `--token <key>` | API key (overrides config) |
| `--json` | Print raw JSON instead of formatted output |
| `-v, --version` | Print version and exit |

---

### `wnp login`

Authenticate with email/password and save the API token.

```bash
wnp login you@example.com
# prompts for password (input is hidden)
```

---

### `wnp config`

```bash
wnp config show                      # print current URL and token prefix
wnp config set-url <url>             # set API base URL
wnp config set-token <token>         # set API key
```

---

### `wnp blueprints` (alias: `bp`)

```bash
wnp bp list                          # list all blueprints
wnp bp get <slug>                    # get a blueprint
wnp bp create <slug> <name>          # create a blueprint
```

---

### `wnp entities` (alias: `ent`, `e`)

```bash
wnp ent list <blueprint-slug>                    # list entities in a blueprint
wnp ent get <blueprint-slug> <entity-slug>       # get an entity
wnp ent create <blueprint-slug> <slug> <name>    # create an entity
wnp ent update <blueprint-slug> <entity-slug>    # update an entity
wnp ent delete <blueprint-slug> <entity-slug>    # delete an entity
```

---

### `wnp checks` (alias: `chk`)

```bash
wnp chk list                         # list all checks
wnp chk list --entity <id-prefix>    # filter by entity ID
wnp chk get <id>                     # get a check
wnp chk results <id>                 # show recent results (default 50)
wnp chk results <id> --limit 200     # show more results
wnp chk transitions <id>             # state-change history for a check
```

**Create a check**

```bash
# HTTP check (default)
wnp chk create --entity <id> --url https://api.example.com/health \
  --interval 60 --status 200 --threshold 3 --timeout 5000

# TCP reachability
wnp chk create --type tcp --entity <id> --url db.internal:5432

# TLS certificate validity
wnp chk create --type tls --entity <id> --url api.example.com:443

# DNS resolution
wnp chk create --type dns --entity <id> --url example.com

# With expected body substring
wnp chk create --entity <id> --url https://api.example.com/health \
  --body '"status":"ok"'
```

**Manage a check**

```bash
wnp chk enable  <id>                 # resume polling
wnp chk disable <id>                 # pause polling (keeps config)
wnp chk update  <id> --interval 30   # change any field
wnp chk delete  <id>                 # delete permanently
```

**Add an alert (shorthand)**

Creates the notification policy if it does not exist, then adds the channel in one step:

```bash
wnp chk alert <id> --type slack   --config url=https://hooks.slack.com/services/...
wnp chk alert <id> --type discord --config url=https://discord.com/api/webhooks/...
wnp chk alert <id> --type webhook --config url=https://example.com/hook
wnp chk alert <id> --type nats    --config subject=alerts.prod
```

---

### `wnp notify` (alias: `n`)

Lower-level notification management:

```bash
wnp notify policy <check-id>         # show policy and channels for a check
wnp notify set-policy <check-id> \
  --cooldown 300 \
  --on-recovery \
  --repeat 3600                      # re-alert every hour while down

wnp notify add-channel <check-id> \
  --type slack --config url=https://hooks.slack.com/services/...

wnp notify rm-channel  <check-id> <channel-id>

wnp notify channels                  # all channels across all checks
wnp notify logs                      # recent delivery log
```

---

### `wnp status`

```bash
wnp status show                      # fleet health summary (up / degraded / down counts)
wnp status transitions               # recent state-change events across all entities
```

---

### `wnp keys`

```bash
wnp keys list
wnp keys create <label>
wnp keys delete <id>
```

---

### `wnp team`

```bash
wnp team list                        # members + pending invitations
wnp team invite <email> --role member|admin
wnp team remove <user-id>
```

---

### `wnp skills`

```bash
wnp skills list
wnp skills get <slug>
wnp skills create <slug> <name>
wnp skills delete <slug>
```

---

## Examples

**Register a service and add monitoring in one shot**

```bash
ENTITY_ID=$(wnp ent create services payments-api "Payments API" --json | jq -r .ID)

wnp chk create --entity $ENTITY_ID \
  --url https://api.example.com/health \
  --interval 60 --status 200 --threshold 3

CHECK_ID=$(wnp chk list --entity $ENTITY_ID --json | jq -r '.[0].ID')

wnp chk alert $CHECK_ID --type slack \
  --config url=https://hooks.slack.com/services/T.../...
```

**Bulk-disable all checks for a deployment window**

```bash
wnp chk list --json | jq -r '.[].ID' | xargs -I{} wnp chk disable {}
# ... deploy ...
wnp chk list --json | jq -r '.[].ID' | xargs -I{} wnp chk enable {}
```

**Watch results for a check**

```bash
watch -n10 wnp chk results <id>
```

**CI/CD — create a check on deploy**

```bash
wnp --token $WNP_API_KEY \
  chk create --entity $ENTITY_ID \
  --url $DEPLOY_URL/health \
  --interval 30 --status 200
```

---

## Free tier limits

New accounts start on the Free plan:

| Resource | Free | Starter ($19/mo) | Pro ($79/mo) |
|----------|------|-----------------|--------------|
| Blueprints | 3 | 15 | 50 |
| Entities per blueprint | 5 | 25 | 100 |
| Checks per entity | 10 | 50 | 200 |
| Result retention | 7 days | 30 days | 90 days |

The API returns a clear error (`LIMIT_EXCEEDED` / `CHECK_LIMIT_EXCEEDED`) when a limit is reached.

---

## Output format

Every command prints a human-readable table by default. Pass `--json` to get raw JSON suitable for piping to `jq`:

```bash
wnp chk list --json | jq '.[] | select(.CheckType=="http") | .ID'
```

---

## License

MIT
