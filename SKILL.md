---
name: robotx
description: Use the robotx CLI to deploy, update, and check status for RobotX applications.
metadata:
  short-description: RobotX deployment CLI skill
---

# RobotX Deployment Skill

This document describes how to integrate RobotX deployment capabilities as a skill for AI agents.

## Overview

RobotX provides a CLI tool that AI agents can use to deploy applications to the RobotX platform. This enables agents to:

1. Create and deploy new projects
2. Update existing projects with new code
3. Manage preview and production environments
4. Monitor build status and logs

## Quick start

- Check CLI availability: `which robotx || which robotx_cli`
- Install if missing: `go install github.com/haibingtown/robotx_cli@latest`
- Ensure PATH includes Go bin: `export PATH="$(go env GOPATH)/bin:$PATH"`
- Optional: symlink `robotx_cli` to `robotx` if you prefer the shorter name

## Configure

- Use a config file at `~/.robotx.yaml` or set env vars:
  - `ROBOTX_BASE_URL`
  - `ROBOTX_API_KEY`

## Deploy (new project)

```bash
robotx deploy [path] --name "My App" [--publish]
```

Parameters (common):
- `--base-url`, `--api-key`
- `--publish` (publish after a successful build)
- `--wait`, `--timeout`

Local-build parameters:
- `--local-build` (build locally and upload artifacts)
- `--install-command` (override install)
- `--build-command` (override build)
- `--output-dir` (override output dir; default from build plan or `dist`)

## Update (existing project)

```bash
robotx update [path] --project-id proj_123 [--publish]
```

The same local-build flags apply to `update`.

## Local build mode

- Build locally, zip the output directory, and upload artifacts to the backend.
- Require backend support for `POST /api/builds/:buildID/artifacts`.
- Use when you want deterministic local builds or when remote build endpoints are unavailable.

## Status and logs

```bash
robotx status --project-id proj_123 --build-id build_456 --logs
```

## Publish

```bash
robotx publish --project-id proj_123 --build-id build_456
```

## Troubleshooting

- `robotx` not found: ensure `$(go env GOPATH)/bin` is on PATH or use `robotx_cli`.
- Build trigger 404: backend API may be older; update backend or use `--local-build`.
