# AI Agent 集成指南

本指南介绍如何将 RobotX CLI 集成到各种 AI agents 中，使它们能够自动部署应用到 RobotX 平台。

## 目录

- [集成方式概览](#集成方式概览)
- [方式 1: 直接 CLI 调用](#方式-1-直接-cli-调用)
- [方式 2: MCP 集成](#方式-2-mcp-集成)
- [方式 3: REST API](#方式-3-rest-api)
- [方式 4: 自定义 Skill](#方式-4-自定义-skill)
- [最佳实践](#最佳实践)

## 集成方式概览

| 方式 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| 直接 CLI 调用 | 任何支持命令行的 agent | 简单、直接 | 需要解析 JSON 输出 |
| MCP 集成 | Claude Desktop 等支持 MCP 的工具 | 原生集成、体验好 | 仅限支持 MCP 的工具 |
| REST API | 需要远程调用的场景 | 跨平台、语言无关 | 需要额外的 API 服务器 |
| 自定义 Skill | 支持 skill 系统的 agent | 配置化、易维护 | 需要 agent 支持 skill |

## 方式 1: 直接 CLI 调用

这是最简单、最通用的集成方式。AI agent 直接调用 `robotx` 命令并解析 JSON 输出。

### Python 示例

```python
import subprocess
import json
from typing import Dict, Any

class RobotXClient:
    """RobotX CLI wrapper for Python"""

    def __init__(self, base_url: str = None, api_key: str = None):
        self.base_url = base_url
        self.api_key = api_key

    def _run_command(self, args: list) -> Dict[str, Any]:
        """Run robotx command and return parsed JSON output"""
        cmd = ['robotx'] + args

        # Add global flags if provided
        if self.base_url:
            cmd.extend(['--base-url', self.base_url])
        if self.api_key:
            cmd.extend(['--api-key', self.api_key])

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            return json.loads(result.stdout)
        else:
            error = json.loads(result.stderr)
            raise Exception(f"RobotX command failed: {error['error']}")

    def deploy(self, project_path: str, name: str = None,
               project_id: str = None, publish: bool = False,
               wait: bool = True, timeout: int = 600) -> Dict[str, Any]:
        """Deploy a project to RobotX"""
        args = ['deploy', project_path]

        if name:
            args.extend(['--name', name])
        if project_id:
            args.extend(['--project-id', project_id])
        if publish:
            args.append('--publish')
        if not wait:
            args.append('--wait=false')
        if timeout != 600:
            args.extend(['--timeout', str(timeout)])

        return self._run_command(args)

    def status(self, project_id: str = None, build_id: str = None) -> Dict[str, Any]:
        """Get project or build status"""
        args = ['status']

        if project_id:
            args.extend(['--project-id', project_id])
        if build_id:
            args.extend(['--build-id', build_id])

        return self._run_command(args)

    def logs(self, build_id: str) -> str:
        """Get build logs"""
        result = self._run_command(['logs', build_id])
        return result.get('logs', '')

    def publish(self, build_id: str) -> Dict[str, Any]:
        """Publish a build to production"""
        return self._run_command(['publish', build_id])

# 使用示例
if __name__ == '__main__':
    # 初始化客户端
    client = RobotXClient(
        base_url='https://api.robotx.xin',
        api_key='your-api-key'
    )

    # 部署项目
    result = client.deploy(
        project_path='./my-app',
        name='my-app',
        publish=True
    )

    print(f"✅ Deployed successfully!")
    print(f"📦 Project ID: {result['project_id']}")
    print(f"🔨 Build ID: {result['build_id']}")
    print(f"🌐 URL: {result['url']}")

    # 查看状态
    status = client.status(project_id=result['project_id'])
    print(f"📊 Status: {status['status']}")

    # 查看日志
    logs = client.logs(result['build_id'])
    print(f"📝 Logs:\n{logs}")
```

### Node.js/TypeScript 示例

```typescript
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface DeployOptions {
  name?: string;
  projectId?: string;
  publish?: boolean;
  wait?: boolean;
  timeout?: number;
}

interface DeployResult {
  success: boolean;
  project_id: string;
  build_id: string;
  status: string;
  url: string;
  message: string;
}

class RobotXClient {
  constructor(
    private baseUrl?: string,
    private apiKey?: string
  ) {}

  private async runCommand(args: string[]): Promise<any> {
    const cmd = ['robotx', ...args];

    if (this.baseUrl) {
      cmd.push('--base-url', this.baseUrl);
    }
    if (this.apiKey) {
      cmd.push('--api-key', this.apiKey);
    }

    try {
      const { stdout } = await execAsync(cmd.join(' '));
      return JSON.parse(stdout);
    } catch (error: any) {
      const errorData = JSON.parse(error.stderr);
      throw new Error(`RobotX command failed: ${errorData.error}`);
    }
  }

  async deploy(
    projectPath: string,
    options: DeployOptions = {}
  ): Promise<DeployResult> {
    const args = ['deploy', projectPath];

    if (options.name) {
      args.push('--name', options.name);
    }
    if (options.projectId) {
      args.push('--project-id', options.projectId);
    }
    if (options.publish) {
      args.push('--publish');
    }
    if (options.wait === false) {
      args.push('--wait=false');
    }
    if (options.timeout) {
      args.push('--timeout', options.timeout.toString());
    }

    return this.runCommand(args);
  }

  async status(projectId?: string, buildId?: string): Promise<any> {
    const args = ['status'];

    if (projectId) {
      args.push('--project-id', projectId);
    }
    if (buildId) {
      args.push('--build-id', buildId);
    }

    return this.runCommand(args);
  }

  async logs(buildId: string): Promise<string> {
    const result = await this.runCommand(['logs', buildId]);
    return result.logs || '';
  }

  async publish(buildId: string): Promise<any> {
    return this.runCommand(['publish', buildId]);
  }
}

// 使用示例
async function main() {
  const client = new RobotXClient(
    'https://api.robotx.xin',
    'your-api-key'
  );

  // 部署项目
  const result = await client.deploy('./my-app', {
    name: 'my-app',
    publish: true
  });

  console.log('✅ Deployed successfully!');
  console.log(`📦 Project ID: ${result.project_id}`);
  console.log(`🔨 Build ID: ${result.build_id}`);
  console.log(`🌐 URL: ${result.url}`);
}

main().catch(console.error);
```

### Go 示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
)

type RobotXClient struct {
    BaseURL string
    APIKey  string
}

type DeployResult struct {
    Success   bool   `json:"success"`
    ProjectID string `json:"project_id"`
    BuildID   string `json:"build_id"`
    Status    string `json:"status"`
    URL       string `json:"url"`
    Message   string `json:"message"`
}

func (c *RobotXClient) runCommand(args []string) (map[string]interface{}, error) {
    cmd := []string{"robotx"}
    cmd = append(cmd, args...)

    if c.BaseURL != "" {
        cmd = append(cmd, "--base-url", c.BaseURL)
    }
    if c.APIKey != "" {
        cmd = append(cmd, "--api-key", c.APIKey)
    }

    output, err := exec.Command(cmd[0], cmd[1:]...).Output()
    if err != nil {
        return nil, fmt.Errorf("command failed: %w", err)
    }

    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse output: %w", err)
    }

    return result, nil
}

func (c *RobotXClient) Deploy(projectPath, name string, publish bool) (*DeployResult, error) {
    args := []string{"deploy", projectPath, "--name", name}
    if publish {
        args = append(args, "--publish")
    }

    result, err := c.runCommand(args)
    if err != nil {
        return nil, err
    }

    data, _ := json.Marshal(result)
    var deployResult DeployResult
    json.Unmarshal(data, &deployResult)

    return &deployResult, nil
}

func main() {
    client := &RobotXClient{
        BaseURL: "https://api.robotx.xin",
        APIKey:  "your-api-key",
    }

    result, err := client.Deploy("./my-app", "my-app", true)
    if err != nil {
        panic(err)
    }

    fmt.Printf("✅ Deployed successfully!\n")
    fmt.Printf("📦 Project ID: %s\n", result.ProjectID)
    fmt.Printf("🔨 Build ID: %s\n", result.BuildID)
    fmt.Printf("🌐 URL: %s\n", result.URL)
}
```

## 方式 2: MCP 集成

Model Context Protocol (MCP) 是一个标准协议，允许 AI 工具与外部服务集成。

### Claude Desktop 集成

1. **配置 MCP 服务器**

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "robotx": {
      "command": "/usr/local/bin/robotx",
      "args": ["mcp"],
      "env": {
        "ROBOTX_BASE_URL": "https://api.robotx.xin",
        "ROBOTX_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

2. **重启 Claude Desktop**

3. **使用示例**

在 Claude Desktop 中，你可以直接使用自然语言：

```
User: 请帮我部署 /path/to/my-app 到 RobotX，项目名称是 my-awesome-app

Claude: 好的，我会使用 RobotX 工具来部署你的项目。

[Claude 会自动调用 RobotX MCP 工具]

✅ 部署成功！
- 项目 ID: proj_abc123
- 构建 ID: build_xyz789
- URL: https://my-awesome-app.api.robotx.xin

你的应用已经成功部署并发布到生产环境。
```

### 可用的 MCP 工具

RobotX MCP 服务器提供以下工具：

- `robotx_deploy`: 部署项目
- `robotx_status`: 查询状态
- `robotx_logs`: 查看日志
- `robotx_publish`: 发布到生产环境
- `robotx_update`: 更新项目配置

## 方式 3: REST API

如果你需要远程调用或跨语言集成，可以使用 REST API 方式。

### 启动 API 服务器（即将支持）

```bash
robotx serve --port 8080
```

### API 端点

```
POST   /api/v1/deploy      - 部署项目
GET    /api/v1/status      - 查询状态
GET    /api/v1/logs/:id    - 查看日志
POST   /api/v1/publish/:id - 发布到生产环境
PUT    /api/v1/projects/:id - 更新项目
```

### 使用示例

```bash
# 部署项目
curl -X POST https://api.api.robotx.xin/v1/deploy \
  -H "Authorization: Bearer your-api-key" \
  -F "project=@project.zip" \
  -F "name=my-app" \
  -F "publish=true"

# 查询状态
curl https://api.api.robotx.xin/v1/status?project_id=proj_123 \
  -H "Authorization: Bearer your-api-key"
```

## 方式 4: 自定义 Skill

对于支持 skill 系统的 AI agents，可以创建一个 RobotX skill。

### Skill 定义示例

```yaml
name: robotx-deploy
version: 1.0.0
description: Deploy applications to RobotX platform

commands:
  deploy:
    description: Deploy a project to RobotX
    command: robotx deploy {project_path} --name {project_name}
    parameters:
      - name: project_path
        type: string
        required: true
        description: Path to the project directory
      - name: project_name
        type: string
        required: true
        description: Name for the project
      - name: publish
        type: boolean
        required: false
        default: false
        description: Publish to production after build
    output_format: json

  status:
    description: Check deployment status
    command: robotx status --project-id {project_id}
    parameters:
      - name: project_id
        type: string
        required: true
        description: Project ID to check
    output_format: json

  logs:
    description: View build logs
    command: robotx logs {build_id}
    parameters:
      - name: build_id
        type: string
        required: true
        description: Build ID to view logs for
    output_format: text
```

## 最佳实践

### 1. 错误处理

始终检查命令的退出码和错误输出：

```python
try:
    result = client.deploy('./my-app', name='my-app')
    print(f"Success: {result['url']}")
except Exception as e:
    print(f"Deployment failed: {e}")
    # 处理错误，可能需要重试或通知用户
```

### 2. 异步部署

对于大型项目，使用异步部署避免阻塞：

```python
# 启动部署但不等待完成
result = client.deploy('./my-app', name='my-app', wait=False)
build_id = result['build_id']

# 稍后检查状态
status = client.status(build_id=build_id)
while status['status'] == 'building':
    time.sleep(5)
    status = client.status(build_id=build_id)

print(f"Build completed: {status['status']}")
```

### 3. 配置管理

使用配置文件或环境变量管理凭证，不要硬编码：

```python
import os

client = RobotXClient(
    base_url=os.getenv('ROBOTX_BASE_URL'),
    api_key=os.getenv('ROBOTX_API_KEY')
)
```

### 4. 日志记录

记录所有部署操作以便调试：

```python
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def deploy_with_logging(project_path, name):
    logger.info(f"Starting deployment: {name}")
    try:
        result = client.deploy(project_path, name=name)
        logger.info(f"Deployment successful: {result['url']}")
        return result
    except Exception as e:
        logger.error(f"Deployment failed: {e}")
        raise
```

### 5. 重试机制

对于网络错误，实现重试机制：

```python
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10)
)
def deploy_with_retry(project_path, name):
    return client.deploy(project_path, name=name)
```

## 示例：完整的 AI Agent 集成

这是一个完整的示例，展示如何在 AI agent 中集成 RobotX：

```python
import os
import logging
from typing import Optional
from robotx_client import RobotXClient

class AIAgentWithRobotX:
    """AI Agent with RobotX deployment capability"""

    def __init__(self):
        self.robotx = RobotXClient(
            base_url=os.getenv('ROBOTX_BASE_URL'),
            api_key=os.getenv('ROBOTX_API_KEY')
        )
        self.logger = logging.getLogger(__name__)

    def create_and_deploy_app(self,
                             app_type: str,
                             app_name: str,
                             requirements: dict) -> dict:
        """
        Create an application based on requirements and deploy it

        Args:
            app_type: Type of application (web, api, etc.)
            app_name: Name for the application
            requirements: Application requirements

        Returns:
            Deployment result with URL and IDs
        """
        # 1. Generate application code
        self.logger.info(f"Generating {app_type} application: {app_name}")
        project_path = self._generate_app_code(app_type, requirements)

        # 2. Deploy to RobotX
        self.logger.info(f"Deploying to RobotX...")
        result = self.robotx.deploy(
            project_path=project_path,
            name=app_name,
            publish=True
        )

        # 3. Verify deployment
        self.logger.info(f"Verifying deployment...")
        status = self.robotx.status(build_id=result['build_id'])

        if status['status'] == 'success':
            self.logger.info(f"✅ Deployment successful: {result['url']}")
            return result
        else:
            raise Exception(f"Deployment failed: {status['status']}")

    def _generate_app_code(self, app_type: str, requirements: dict) -> str:
        """Generate application code based on type and requirements"""
        # Your code generation logic here
        pass

# 使用示例
if __name__ == '__main__':
    agent = AIAgentWithRobotX()

    result = agent.create_and_deploy_app(
        app_type='web',
        app_name='my-awesome-app',
        requirements={
            'framework': 'express',
            'features': ['api', 'auth', 'database']
        }
    )

    print(f"🎉 App deployed: {result['url']}")
```

## 故障排查

### 常见问题

1. **命令未找到**
   ```bash
   # 确保 robotx 在 PATH 中
   which robotx
   # 或使用完整路径
   /usr/local/bin/robotx --version
   ```

2. **认证失败**
   ```bash
   # 检查 API key 是否正确
   robotx status --project-id test --api-key your-key
   ```

3. **JSON 解析错误**
   ```python
   # 确保使用正确的输出流
   result = subprocess.run(cmd, capture_output=True, text=True)
   output = result.stdout  # 不是 result.stderr
   ```

## 更多资源

- [RobotX CLI 文档](README.md)
- [API 参考](API.md)
- [示例项目](examples/)
- [常见问题](FAQ.md)
