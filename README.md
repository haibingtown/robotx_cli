# RobotX CLI

RobotX CLI 是一个命令行工具，用于将 AI 应用部署到 RobotX 平台。它为 AI agents（如 Claude Code、Cursor 等）提供了简单的接口来创建、构建和部署项目。

## 功能特性

- 🚀 **一键部署**: 自动打包、上传、构建和发布项目
- 📊 **状态查询**: 实时查看项目和构建状态
- 📝 **日志查看**: 获取构建和运行日志
- 🔄 **项目更新**: 更新现有项目配置
- 🤖 **MCP 集成**: 支持 Model Context Protocol，可与 Claude Desktop 集成
- 🎯 **AI Agent 友好**: JSON 输出格式，易于程序解析

## 安装

### 从源码构建

```bash
cd cli
make build
make install
```

### 直接使用二进制

```bash
# 构建
cd cli
go build -o robotx cmd/robotx/main.go

# 使用
./robotx --help
```

## 配置

### 方式 1: 配置文件

创建 `~/.robotx.yaml`:

```yaml
# RobotX Server base URL
base_url: https://api.robotx.xin

# API Key for authentication
api_key: your-api-key-here

# Optional: Default project visibility
default_visibility: private

# Optional: Default build timeout in seconds
default_timeout: 600
```

### 方式 2: 环境变量

```bash
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=your-api-key-here
```

### 方式 3: 命令行参数

```bash
robotx deploy /path/to/project \
  --base-url https://api.robotx.xin \
  --api-key your-api-key-here \
  --name my-project
```

## 使用方法

### 1. 部署项目

```bash
# 部署新项目
robotx deploy /path/to/project --name my-app

# 部署并发布到生产环境
robotx deploy /path/to/project --name my-app --publish

# 更新现有项目
robotx deploy /path/to/project --project-id proj_123456

# 不等待构建完成
robotx deploy /path/to/project --name my-app --wait=false
```

**输出示例**:
```json
{
  "success": true,
  "project_id": "proj_123456",
  "build_id": "build_789012",
  "status": "success",
  "url": "https://my-app.api.robotx.xin",
  "message": "Deployment completed successfully"
}
```

### 2. 查询状态

```bash
# 查询项目状态
robotx status --project-id proj_123456

# 查询构建状态
robotx status --build-id build_789012
```

**输出示例**:
```json
{
  "success": true,
  "status": "running",
  "project": {
    "id": "proj_123456",
    "name": "my-app",
    "visibility": "private"
  },
  "build": {
    "id": "build_789012",
    "status": "success",
    "created_at": "2024-02-03T16:00:00Z"
  }
}
```

### 3. 查看日志

```bash
# 查看构建日志
robotx logs build_789012

# 实时跟踪日志（即将支持）
robotx logs build_789012 --follow
```

### 4. 更新项目

```bash
# 更新项目配置
robotx update proj_123456 \
  --name new-name \
  --visibility public
```

### 5. 发布到生产环境

```bash
# 发布指定构建
robotx publish build_789012
```

## AI Agent 集成

### 集成方式 1: 直接调用 CLI

AI agents 可以直接调用 `robotx` 命令并解析 JSON 输出：

```python
import subprocess
import json

def deploy_to_robotx(project_path, project_name):
    """Deploy a project using RobotX CLI"""
    result = subprocess.run(
        ['robotx', 'deploy', project_path, '--name', project_name],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return json.loads(result.stdout)
    else:
        error = json.loads(result.stderr)
        raise Exception(f"Deployment failed: {error['error']}")

# 使用示例
response = deploy_to_robotx('/path/to/project', 'my-app')
print(f"Deployed to: {response['url']}")
```

### 集成方式 2: MCP (Model Context Protocol)

RobotX CLI 支持作为 MCP 服务器运行，可与 Claude Desktop 等工具集成。

#### 配置 Claude Desktop

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

重启 Claude Desktop 后，你可以直接在对话中使用 RobotX 功能：

```
User: 请帮我部署这个项目到 RobotX
Claude: 好的，我会使用 RobotX 工具来部署你的项目...
```

### 集成方式 3: 作为 Skill

对于支持自定义 skills 的 AI agents，可以将 RobotX CLI 封装为一个 skill：

**skill.yaml** (示例):
```yaml
name: robotx-deploy
description: Deploy applications to RobotX platform
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
output_format: json
```

## 项目要求

要部署到 RobotX，你的项目需要包含以下文件之一：

### Node.js 项目
- `package.json` (必需)
- `Dockerfile` (可选，如果没有会自动生成)

### Python 项目
- `requirements.txt` 或 `pyproject.toml` (必需)
- `Dockerfile` (可选)

### Go 项目
- `go.mod` (必需)
- `Dockerfile` (可选)

### 通用项目
- `Dockerfile` (必需)

## 完整示例

### 示例 1: 部署 Express.js 应用

```bash
# 1. 创建项目
mkdir my-express-app
cd my-express-app

# 2. 创建 package.json
cat > package.json << EOF
{
  "name": "my-express-app",
  "version": "1.0.0",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  }
}
EOF

# 3. 创建应用代码
cat > index.js << EOF
const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello from RobotX!' });
});

app.listen(port, () => {
  console.log(\`Server running on port \${port}\`);
});
EOF

# 4. 部署到 RobotX
robotx deploy . --name my-express-app --publish
```

### 示例 2: 部署 Python FastAPI 应用

```bash
# 1. 创建项目
mkdir my-fastapi-app
cd my-fastapi-app

# 2. 创建 requirements.txt
cat > requirements.txt << EOF
fastapi==0.104.1
uvicorn==0.24.0
EOF

# 3. 创建应用代码
cat > main.py << EOF
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Hello from RobotX!"}
EOF

# 4. 创建 Dockerfile
cat > Dockerfile << EOF
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
EOF

# 5. 部署到 RobotX
robotx deploy . --name my-fastapi-app --publish
```

## 错误处理

所有命令在失败时会返回非零退出码，并输出 JSON 格式的错误信息：

```json
{
  "success": false,
  "error": "Project not found",
  "details": "No project found with ID: proj_123456"
}
```

常见错误码：
- `1`: 一般错误（配置错误、参数错误等）
- `2`: API 错误（认证失败、网络错误等）
- `3`: 构建失败
- `4`: 部署失败

## 高级用法

### 自定义构建超时

```bash
robotx deploy /path/to/project \
  --name my-app \
  --timeout 1200  # 20 分钟
```

### 设置项目可见性

```bash
robotx deploy /path/to/project \
  --name my-app \
  --visibility public
```

### 环境变量传递（即将支持）

```bash
robotx deploy /path/to/project \
  --name my-app \
  --env NODE_ENV=production \
  --env API_KEY=secret
```

## 开发

### 项目结构

```
cli/
├── cmd/
│   └── robotx/
│       └── main.go          # CLI 入口点
├── internal/
│   ├── client/
│   │   └── client.go        # RobotX API 客户端
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── mcp/
│   │   └── server.go        # MCP 服务器实现
│   └── output/
│       └── output.go        # 输出格式化
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 运行测试

```bash
make test
```

### 构建所有平台

```bash
make build-all
```

## 贡献

欢迎提交 issues 和 pull requests！

## 许可证

MIT License

## 相关链接

- [RobotX Server](../server/)
- [RobotX SDK](../sdk/)
- [RobotX 文档](../docs/)
