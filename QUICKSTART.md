# RobotX CLI 快速开始

这是一个 5 分钟快速入门指南，帮助你快速上手 RobotX CLI。

## 1. 安装 (1 分钟)

```bash
# 克隆仓库
git clone https://github.com/your-org/robotx.git
cd haibingtown/robotx_cli

# 构建并安装
make build
make install

# 验证安装
robotx --version
```

## 2. 配置 (1 分钟)

创建配置文件 `~/.robotx.yaml`:

```bash
cat > ~/.robotx.yaml << EOF
base_url: https://your-robotx-server.com
api_key: your-api-key-here
EOF
```

或者使用环境变量：

```bash
export ROBOTX_BASE_URL=https://your-robotx-server.com
export ROBOTX_API_KEY=your-api-key-here
```

## 3. 部署你的第一个应用 (3 分钟)

### 示例：部署一个 Node.js 应用

```bash
# 创建项目目录
mkdir hello-robotx
cd hello-robotx

# 创建 package.json
cat > package.json << 'EOF'
{
  "name": "hello-robotx",
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

# 创建应用代码
cat > index.js << 'EOF'
const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({
    message: 'Hello from RobotX!',
    timestamp: new Date().toISOString()
  });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
EOF

# 部署到 RobotX
robotx deploy . --name hello-robotx --publish
```

**输出示例**:
```json
{
  "success": true,
  "project_id": "proj_abc123",
  "build_id": "build_xyz789",
  "status": "success",
  "url": "https://hello-robotx.your-domain.com",
  "message": "Deployment completed successfully"
}
```

🎉 恭喜！你的应用已经部署成功！访问输出中的 URL 即可看到你的应用。

## 4. 常用命令

```bash
# 查看项目状态
robotx status --project-id proj_abc123

# 查看构建日志
robotx logs build_xyz789

# 更新项目
robotx deploy . --project-id proj_abc123

# 发布到生产环境
robotx publish build_xyz789
```

## 5. 集成到 AI Agent

### Python 示例

```python
import subprocess
import json

def deploy_with_robotx(project_path, name):
    result = subprocess.run(
        ['robotx', 'deploy', project_path, '--name', name],
        capture_output=True,
        text=True
    )
    return json.loads(result.stdout)

# 使用
response = deploy_with_robotx('./my-app', 'my-app')
print(f"Deployed to: {response['url']}")
```

### Claude Desktop 集成

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "robotx": {
      "command": "/usr/local/bin/robotx",
      "args": ["mcp"],
      "env": {
        "ROBOTX_BASE_URL": "https://your-robotx-server.com",
        "ROBOTX_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

重启 Claude Desktop，然后你就可以在对话中直接使用 RobotX 功能了！

## 下一步

- 📖 阅读[完整文档](README.md)了解更多功能
- 🔧 查看[示例项目](examples/)
- 💬 加入我们的[社区讨论](https://github.com/your-org/robotx/discussions)

## 需要帮助？

- 查看 [FAQ](docs/FAQ.md)
- 提交 [Issue](https://github.com/your-org/robotx/issues)
- 查看 [API 文档](docs/API.md)
