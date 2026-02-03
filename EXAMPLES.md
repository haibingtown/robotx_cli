# RobotX Deployment Skill - Integration Examples

This document provides practical examples of how AI agents can integrate with RobotX deployment capabilities.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Claude Code Integration](#claude-code-integration)
3. [Cursor Integration](#cursor-integration)
4. [Generic AI Agent Integration](#generic-ai-agent-integration)
5. [Advanced Patterns](#advanced-patterns)

## Quick Start

### Prerequisites

```bash
# Install RobotX CLI
cd cli
make build
make install

# Configure credentials
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=your-api-key-here

# Verify installation
robotx --help
```

### Basic Deployment

```bash
# Deploy a new project
robotx deploy ./my-app --name "My App"

# Output will include:
# - Project ID: proj_xxx
# - Build ID: build_xxx
# - Preview URL: https://api.robotx.xin/preview/proj_xxx
```

## Claude Code Integration

### Example 1: Deploy Generated Code

```typescript
// Claude Code can execute bash commands
async function deployToRobotX(projectPath: string, projectName: string) {
  const command = `robotx deploy ${projectPath} --name "${projectName}"`;

  const result = await executeBash(command);

  // Parse output
  const projectIdMatch = result.stdout.match(/Project created: (proj_\w+)/);
  const previewUrlMatch = result.stdout.match(/Preview URL: (https:\/\/[^\s]+)/);

  return {
    projectId: projectIdMatch?.[1],
    previewUrl: previewUrlMatch?.[1],
    success: result.exitCode === 0
  };
}

// Usage
const deployment = await deployToRobotX('./generated-app', 'AI Generated App');
console.log(`Deployed to: ${deployment.previewUrl}`);
```

### Example 2: Iterative Development

```typescript
async function updateProject(projectId: string, projectPath: string) {
  const command = `robotx update ${projectPath} --project-id ${projectId}`;

  const result = await executeBash(command);

  // Parse build ID
  const buildIdMatch = result.stdout.match(/Build started: (build_\w+)/);

  return {
    buildId: buildIdMatch?.[1],
    success: result.exitCode === 0
  };
}

// Usage in iterative development
let projectId = 'proj_abc123';

// Make changes
await modifyCode('./my-app/src/index.ts');

// Deploy update
const update = await updateProject(projectId, './my-app');
console.log(`Build ID: ${update.buildId}`);
```

### Example 3: Status Monitoring

```typescript
async function waitForBuild(projectId: string, buildId: string): Promise<boolean> {
  const maxAttempts = 60; // 5 minutes with 5s intervals

  for (let i = 0; i < maxAttempts; i++) {
    const command = `robotx status --project-id ${projectId} --build-id ${buildId}`;
    const result = await executeBash(command);

    if (result.stdout.includes('Status: success')) {
      return true;
    }

    if (result.stdout.includes('Status: failed')) {
      // Get logs
      const logsCommand = `robotx status --project-id ${projectId} --build-id ${buildId} --logs`;
      const logsResult = await executeBash(logsCommand);
      console.error('Build failed:', logsResult.stdout);
      return false;
    }

    // Wait 5 seconds
    await new Promise(resolve => setTimeout(resolve, 5000));
  }

  throw new Error('Build timeout');
}
```

## Cursor Integration

### Example 1: Python Integration

```python
import subprocess
import re
import time

def deploy_to_robotx(project_path: str, project_name: str) -> dict:
    """Deploy a project to RobotX"""

    cmd = ['robotx', 'deploy', project_path, '--name', project_name]
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode != 0:
        raise Exception(f"Deployment failed: {result.stderr}")

    # Parse output
    project_id = re.search(r'Project created: (proj_\w+)', result.stdout)
    preview_url = re.search(r'Preview URL: (https://[^\s]+)', result.stdout)

    return {
        'project_id': project_id.group(1) if project_id else None,
        'preview_url': preview_url.group(1) if preview_url else None,
        'output': result.stdout
    }

# Usage
deployment = deploy_to_robotx('./my-app', 'My AI App')
print(f"Deployed to: {deployment['preview_url']}")
```

### Example 2: Update and Publish

```python
def update_and_publish(project_id: str, project_path: str) -> dict:
    """Update project and publish to production"""

    # Update
    update_cmd = ['robotx', 'update', project_path,
                  '--project-id', project_id, '--publish']
    result = subprocess.run(update_cmd, capture_output=True, text=True)

    if result.returncode != 0:
        raise Exception(f"Update failed: {result.stderr}")

    # Parse output
    build_id = re.search(r'Build started: (build_\w+)', result.stdout)
    prod_url = re.search(r'Production URL: (https://[^\s]+)', result.stdout)

    return {
        'build_id': build_id.group(1) if build_id else None,
        'production_url': prod_url.group(1) if prod_url else None,
        'output': result.stdout
    }

# Usage
result = update_and_publish('proj_abc123', './my-app')
print(f"Published to: {result['production_url']}")
```

### Example 3: Status Check with Retry

```python
def check_build_status(project_id: str, build_id: str,
                       max_retries: int = 60) -> str:
    """Check build status with retry"""

    for i in range(max_retries):
        cmd = ['robotx', 'status',
               '--project-id', project_id,
               '--build-id', build_id]
        result = subprocess.run(cmd, capture_output=True, text=True)

        if 'Status: success' in result.stdout:
            return 'success'

        if 'Status: failed' in result.stdout:
            # Get logs
            logs_cmd = cmd + ['--logs']
            logs_result = subprocess.run(logs_cmd, capture_output=True, text=True)
            raise Exception(f"Build failed:\n{logs_result.stdout}")

        time.sleep(5)

    raise Exception('Build timeout')

# Usage
status = check_build_status('proj_abc123', 'build_xyz789')
print(f"Build status: {status}")
```

## Generic AI Agent Integration

### Bash Script Integration

```bash
#!/bin/bash

# Function to deploy project
deploy_project() {
    local project_path=$1
    local project_name=$2

    echo "Deploying $project_name from $project_path..."

    output=$(robotx deploy "$project_path" --name "$project_name" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        project_id=$(echo "$output" | grep -oP 'Project created: \K\w+')
        preview_url=$(echo "$output" | grep -oP 'Preview URL: \K\S+')

        echo "Success!"
        echo "Project ID: $project_id"
        echo "Preview URL: $preview_url"

        return 0
    else
        echo "Deployment failed:"
        echo "$output"
        return 1
    fi
}

# Function to update project
update_project() {
    local project_id=$1
    local project_path=$2

    echo "Updating project $project_id..."

    output=$(robotx update "$project_path" --project-id "$project_id" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        build_id=$(echo "$output" | grep -oP 'Build started: \K\w+')
        echo "Success! Build ID: $build_id"
        return 0
    else
        echo "Update failed:"
        echo "$output"
        return 1
    fi
}

# Usage
deploy_project "./my-app" "My App"
```

### Node.js Integration

```javascript
const { exec } = require('child_process');
const util = require('util');
const execPromise = util.promisify(exec);

async function deployToRobotX(projectPath, projectName) {
  try {
    const { stdout, stderr } = await execPromise(
      `robotx deploy ${projectPath} --name "${projectName}"`
    );

    // Parse output
    const projectIdMatch = stdout.match(/Project created: (proj_\w+)/);
    const previewUrlMatch = stdout.match(/Preview URL: (https:\/\/[^\s]+)/);

    return {
      success: true,
      projectId: projectIdMatch?.[1],
      previewUrl: previewUrlMatch?.[1],
      output: stdout
    };
  } catch (error) {
    return {
      success: false,
      error: error.message,
      output: error.stdout || error.stderr
    };
  }
}

// Usage
(async () => {
  const result = await deployToRobotX('./my-app', 'My App');
  if (result.success) {
    console.log(`Deployed to: ${result.previewUrl}`);
  } else {
    console.error(`Deployment failed: ${result.error}`);
  }
})();
```

## Advanced Patterns

### Pattern 1: Multi-Environment Deployment

```bash
#!/bin/bash

# Deploy to preview
deploy_preview() {
    robotx deploy . --name "$1"
}

# Test in preview
test_preview() {
    local preview_url=$1
    # Run tests against preview URL
    curl -f "$preview_url" || return 1
}

# Promote to production
promote_to_production() {
    local project_id=$1
    local build_id=$2
    robotx publish --project-id "$project_id" --build-id "$build_id"
}

# Full workflow
project_name="My App"
deployment=$(deploy_preview "$project_name")
project_id=$(echo "$deployment" | grep -oP 'Project created: \K\w+')
build_id=$(echo "$deployment" | grep -oP 'Build started: \K\w+')
preview_url=$(echo "$deployment" | grep -oP 'Preview URL: \K\S+')

if test_preview "$preview_url"; then
    echo "Tests passed, promoting to production..."
    promote_to_production "$project_id" "$build_id"
else
    echo "Tests failed, not promoting"
    exit 1
fi
```

### Pattern 2: Continuous Deployment

```python
import subprocess
import time
import hashlib
import os

def get_directory_hash(path: str) -> str:
    """Calculate hash of directory contents"""
    hasher = hashlib.md5()
    for root, dirs, files in os.walk(path):
        for file in sorted(files):
            filepath = os.path.join(root, file)
            with open(filepath, 'rb') as f:
                hasher.update(f.read())
    return hasher.hexdigest()

def watch_and_deploy(project_path: str, project_id: str, interval: int = 30):
    """Watch directory and auto-deploy on changes"""

    last_hash = get_directory_hash(project_path)
    print(f"Watching {project_path} for changes...")

    while True:
        time.sleep(interval)

        current_hash = get_directory_hash(project_path)
        if current_hash != last_hash:
            print("Changes detected, deploying...")

            cmd = ['robotx', 'update', project_path,
                   '--project-id', project_id]
            result = subprocess.run(cmd, capture_output=True, text=True)

            if result.returncode == 0:
                print("Deployment successful!")
            else:
                print(f"Deployment failed: {result.stderr}")

            last_hash = current_hash

# Usage
watch_and_deploy('./my-app', 'proj_abc123', interval=30)
```

### Pattern 3: Rollback on Failure

```python
def deploy_with_rollback(project_id: str, project_path: str,
                         previous_build_id: str) -> dict:
    """Deploy with automatic rollback on failure"""

    # Deploy update
    update_cmd = ['robotx', 'update', project_path,
                  '--project-id', project_id]
    result = subprocess.run(update_cmd, capture_output=True, text=True)

    if result.returncode != 0:
        raise Exception(f"Update failed: {result.stderr}")

    # Get new build ID
    build_id = re.search(r'Build started: (build_\w+)', result.stdout)
    new_build_id = build_id.group(1) if build_id else None

    # Check build status
    try:
        status = check_build_status(project_id, new_build_id)

        if status == 'success':
            # Publish new build
            publish_cmd = ['robotx', 'publish',
                          '--project-id', project_id,
                          '--build-id', new_build_id]
            subprocess.run(publish_cmd, check=True)

            return {
                'success': True,
                'build_id': new_build_id
            }
    except Exception as e:
        # Rollback to previous build
        print(f"Deployment failed, rolling back to {previous_build_id}...")
        rollback_cmd = ['robotx', 'publish',
                       '--project-id', project_id,
                       '--build-id', previous_build_id]
        subprocess.run(rollback_cmd, check=True)

        return {
            'success': False,
            'error': str(e),
            'rolled_back_to': previous_build_id
        }

# Usage
result = deploy_with_rollback('proj_abc123', './my-app', 'build_old456')
if result['success']:
    print(f"Deployed successfully: {result['build_id']}")
else:
    print(f"Deployment failed, rolled back: {result['error']}")
```

## Best Practices

### 1. Error Handling

Always check exit codes and parse error messages:

```bash
if ! robotx deploy . --name "My App"; then
    echo "Deployment failed"
    exit 1
fi
```

### 2. Environment Variables

Use environment variables for credentials:

```bash
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=$SECRET_API_KEY

robotx deploy . --name "My App"
```

### 3. Output Parsing

Use regex to extract structured data:

```python
import re

output = "Project created: proj_abc123"
project_id = re.search(r'proj_\w+', output).group(0)
```

### 4. Timeout Handling

Set appropriate timeouts for long-running operations:

```bash
robotx deploy . --name "My App" --timeout 1200  # 20 minutes
```

### 5. Logging

Capture and log all output for debugging:

```python
result = subprocess.run(cmd, capture_output=True, text=True)
with open('deployment.log', 'a') as f:
    f.write(result.stdout)
    f.write(result.stderr)
```

## Troubleshooting

### Common Issues

1. **CLI not found**
   ```bash
   which robotx
   # If not found, add to PATH or use full path
   ```

2. **Authentication failed**
   ```bash
   # Check credentials
   echo $ROBOTX_API_KEY
   # Verify with status command
   robotx status --project-id test
   ```

3. **Build timeout**
   ```bash
   # Increase timeout
   robotx deploy . --name "My App" --timeout 1200
   ```

4. **Parse errors**
   ```bash
   # Use --verbose for more output
   robotx deploy . --name "My App" --verbose
   ```

## Support

For more information:
- CLI Documentation: [README.md](README.md)
- Skill Documentation: [SKILL.md](SKILL.md)
- API Reference: https://api.api.robotx.xin/docs
