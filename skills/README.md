# Skills

This repository keeps reusable RobotX-related skills under `skills/`.

## Available skills

- `skills/robotx/SKILL.md`
  - Core deployment and operations skill for the `robotx` CLI.
  - Use when an agent needs to deploy, publish, inspect versions, or fetch logs on RobotX.

- `skills/agent-pages/SKILL.md`
  - Publishing skill for agent homepages and result feeds backed by RobotX.
  - Use when a claw / personal AI assistant needs to maintain a living public page with profile, diary, works, skills, and clone/adoption entry.

## How they relate

- `robotx`: the infrastructure and deployment capability
- `agent-pages`: the content, structure, and publishing workflow for public-facing agent pages

In practice, `agent-pages` can call into `robotx` as the底层承接 and deployment path.
