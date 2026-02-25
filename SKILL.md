---
name: robotx
description: Use the robotx CLI to deploy, update, and check status for RobotX applications.
metadata:
  short-description: RobotX deployment CLI skill
---

# RobotX Deployment Skill

Use this skill when an agent needs to deploy or update a project on RobotX using the `robotx` CLI.

## Quick start

- Check CLI availability: `which robotx || which robotx_cli`
- Install (binary-first, no Go required):
  - `curl -fsSL https://raw.githubusercontent.com/haibingtown/robotx_cli/main/scripts/install.sh | bash`
- Fallback install (if you explicitly want source install):
  - `go install github.com/haibingtown/robotx_cli@latest`
  - `export PATH="$(go env GOPATH)/bin:$PATH"`

## Configure

Set credentials by config file (`~/.robotx.yaml`) or env vars:

- `ROBOTX_BASE_URL`
- `ROBOTX_API_KEY`

## Machine-readable output

For agents and workflows, always use structured output:

- `robotx deploy . --name my-app --output json`
- `robotx update . --project-id proj_123 --output json`
- `robotx status --project-id proj_123 --output json`
- `robotx logs --build-id build_456 --output json`
- `robotx publish --project-id proj_123 --build-id build_456 --output json`

JSON is written to stdout. Progress logs are written to stderr.

## Common commands

### Deploy new project

```bash
robotx deploy [path] --name "My App" [--publish] [--wait=true]
```

### Update existing project

```bash
robotx update [path] --project-id proj_123 [--publish]
```

### Status

```bash
robotx status --project-id proj_123 [--build-id build_456] [--logs]
```

`status` accepts `--project-id`, `--build-id`, or both. If `--logs` is set, `--build-id` is required.

### Logs

```bash
robotx logs --build-id build_456 [--project-id proj_123]
```

### Publish

```bash
robotx publish --project-id proj_123 --build-id build_456
```

## MCP note

`robotx mcp` is currently a placeholder and not available for production use. Use shell/CLI mode for agent integration.
