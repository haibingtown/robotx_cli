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

## Architecture

```
┌─────────────────┐
│   AI Agent      │
│ (Claude Code,   │
│  Cursor, etc.)  │
└────────┬────────┘
         │ Execute CLI commands
         ▼
┌─────────────────┐
│  RobotX CLI     │
│  (robotx)       │
└────────┬────────┘
         │ REST API calls
         ▼
┌─────────────────┐
│  RobotX Server  │
│  (API)          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Build & Deploy │
│  (Runtime)      │
└─────────────────┘
```

## Installation

### For AI Agents

AI agents should ensure the RobotX CLI is installed and configured:

```bash
# Check if robotx CLI is available
which robotx

# If not installed, install from the public repo
go install github.com/haibingtown/robotx_cli@latest

# Configure credentials
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=your-api-key
```

## Skill Definition

### Skill: Deploy Application

**Purpose**: Deploy an application to RobotX platform

**Command**:
```bash
robotx deploy [path] --name <project-name> [--publish]
```

**Parameters**:
- `path`: Project directory path (default: current directory)
- `--name`: Project name (required for new projects)
- `--publish`: Publish to production after build (optional)
- `--base-url`: RobotX server URL (or use ROBOTX_BASE_URL env)
- `--api-key`: API key (or use ROBOTX_API_KEY env)

**Returns**:
- Project ID
- Build ID
- Preview URL
- Production URL (if published)

**Example**:
```bash
robotx deploy ./my-app --name "My App" --publish
```

### Skill: Update Application

**Purpose**: Update an existing application with new code

**Command**:
```bash
robotx update [path] --project-id <project-id> [--publish]
```

**Parameters**:
- `path`: Project directory path (default: current directory)
- `--project-id`: Existing project ID (required)
- `--publish`: Publish to production after build (optional)

**Returns**:
- Build ID
- Preview URL
- Production URL (if published)

**Example**:
```bash
robotx update --project-id proj_abc123 --publish
```

### Skill: Check Status

**Purpose**: Check project or build status

**Command**:
```bash
robotx status --project-id <project-id> [--build-id <build-id>] [--logs]
```

**Parameters**:
- `--project-id`: Project ID (required)
- `--build-id`: Build ID (optional)
- `--logs`: Show build logs (optional)

**Returns**:
- Project information
- Build status
- Build logs (if requested)
- URLs

**Example**:
```bash
robotx status --project-id proj_abc123 --build-id build_xyz789 --logs
```

### Skill: Publish to Production

**Purpose**: Publish a specific build to production

**Command**:
```bash
robotx publish --project-id <project-id> --build-id <build-id>
```

**Parameters**:
- `--project-id`: Project ID (required)
- `--build-id`: Build ID (required)

**Returns**:
- Production URL

**Example**:
```bash
robotx publish --project-id proj_abc123 --build-id build_xyz789
```

## Usage Patterns

### Pattern 1: New Project Deployment

```bash
# Agent creates a new application
# ... code generation ...

# Deploy to RobotX
robotx deploy . --name "Generated App"

# Output parsing:
# - Extract project_id from "Project created: proj_xxx"
# - Extract preview URL from "Preview URL: https://..."
```

### Pattern 2: Iterative Development

```bash
# Initial deployment
robotx deploy . --name "My App"
# Save project_id: proj_abc123

# Make changes
# ... code modifications ...

# Update deployment
robotx update --project-id proj_abc123

# Test in preview
# Visit preview URL

# Publish when ready
robotx publish --project-id proj_abc123 --build-id build_xyz789
```

### Pattern 3: Status Monitoring

```bash
# Check if build is complete
robotx status --project-id proj_abc123 --build-id build_xyz789

# Parse output for status:
# - "Status: success" → build complete
# - "Status: running" → still building
# - "Status: failed" → build failed

# Get logs if failed
robotx status --project-id proj_abc123 --build-id build_xyz789 --logs
```

## Integration Examples

### Claude Code Integration

Claude Code can use the RobotX CLI through bash commands:

```typescript
// Example: Deploy a project
const deployCommand = `robotx deploy . --name "${projectName}" --base-url ${baseUrl} --api-key ${apiKey}`;
const result = await executeBash(deployCommand);

// Parse output
const projectIdMatch = result.match(/Project created: (proj_\w+)/);
const projectId = projectIdMatch ? projectIdMatch[1] : null;

const previewUrlMatch = result.match(/Preview URL: (https:\/\/[^\s]+)/);
const previewUrl = previewUrlMatch ? previewUrlMatch[1] : null;
```

### Cursor Integration

Cursor can execute commands and parse output:

```python
import subprocess
import re

# Deploy project
result = subprocess.run(
    ['robotx', 'deploy', '.', '--name', 'My App'],
    capture_output=True,
    text=True
)

# Parse project ID
project_id_match = re.search(r'Project created: (proj_\w+)', result.stdout)
project_id = project_id_match.group(1) if project_id_match else None

# Parse preview URL
preview_url_match = re.search(r'Preview URL: (https://[^\s]+)', result.stdout)
preview_url = preview_url_match.group(1) if preview_url_match else None
```

## Output Parsing

The CLI provides structured output that's easy to parse:

### Success Indicators
- `✅` - Operation completed successfully
- `📦` - Project/packaging operation
- `⬆️` - Upload operation
- `🔨` - Build operation
- `🚀` - Publish operation
- `🌐` - URL information

### Key Patterns
- Project ID: `proj_[a-z0-9]+`
- Build ID: `build_[a-z0-9]+`
- Commit ID: `commit_[a-z0-9]+`
- Preview URL: `https://.*/preview/proj_[a-z0-9]+`
- Production URL: `https://.*/proj_[a-z0-9]+`

### Status Values
- `queued` - Build is queued
- `running` - Build is in progress
- `success` - Build completed successfully
- `failed` - Build failed

## Error Handling

### Common Errors

1. **Missing Configuration**
```
Error: base URL is required (use --base-url or set ROBOTX_BASE_URL)
```
**Solution**: Set ROBOTX_BASE_URL environment variable or use --base-url flag

2. **Authentication Failed**
```
Error: API error: unauthorized
```
**Solution**: Check API key is valid

3. **Build Failed**
```
❌ Build failed with status: failed
```
**Solution**: Check build logs with `--logs` flag

4. **Timeout**
```
Error: build timeout after 600 seconds
```
**Solution**: Increase timeout with `--timeout` flag

## Best Practices

### For AI Agents

1. **Always check CLI availability first**
```bash
if ! command -v robotx &> /dev/null; then
    echo "RobotX CLI not found. Installing..."
    go install github.com/haibingtown/robotx_cli@latest
fi
```

2. **Use environment variables for credentials**
```bash
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=$API_KEY
```

3. **Parse output for IDs and URLs**
```bash
output=$(robotx deploy . --name "My App")
project_id=$(echo "$output" | grep -oP 'Project created: \K\w+')
preview_url=$(echo "$output" | grep -oP 'Preview URL: \K\S+')
```

4. **Handle errors gracefully**
```bash
if robotx deploy . --name "My App"; then
    echo "Deployment successful"
else
    echo "Deployment failed"
    exit 1
fi
```

5. **Use --wait flag for synchronous operations**
```bash
# Wait for build to complete
robotx deploy . --name "My App" --wait

# Don't wait (async)
robotx deploy . --name "My App" --wait=false
```

## Security Considerations

1. **API Key Management**
   - Never hardcode API keys in code
   - Use environment variables or secure vaults
   - Rotate keys regularly

2. **Project Visibility**
   - Use `--visibility private` for sensitive projects
   - Default is private

3. **Access Control**
   - API keys are scoped to specific projects
   - Use separate keys for different environments

## Troubleshooting

### Debug Mode

Enable verbose output:
```bash
robotx deploy . --name "My App" --verbose
```

### Check Configuration

```bash
robotx config show
```

### Test Connection

```bash
robotx status --project-id test
```

## Future Enhancements

Planned features:

1. **MCP Protocol Support** - Native integration with Model Context Protocol
2. **Function Calling** - OpenAI-compatible function definitions
3. **Webhooks** - Real-time build notifications
4. **Rollback** - Quick rollback to previous builds
5. **Environment Variables** - Manage runtime environment variables
6. **Custom Domains** - Configure custom domains for projects

## Support

For issues or questions:
- GitHub Issues: https://github.com/your-org/robotx/issues
- Documentation: https://docs.api.robotx.xin
- API Reference: https://api.api.robotx.xin/docs
