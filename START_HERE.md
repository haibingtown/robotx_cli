# 🚀 从这里开始

欢迎使用 RobotX CLI！这是一个简短的指南，帮助你快速找到需要的信息。

## 📖 我应该读哪个文档？

### 🎯 我想快速了解项目
👉 阅读 [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)（3 分钟）

### ⚡ 我想立即开始使用
👉 阅读 [QUICKSTART.md](QUICKSTART.md)（5 分钟）

### 📚 我想了解完整功能
👉 阅读 [README.md](README.md)（15 分钟）

### 💡 我想看使用示例
👉 阅读 [EXAMPLES.md](EXAMPLES.md)（10 分钟）

### 🤖 我想集成到 AI Agent
👉 阅读 [docs/AI_AGENT_INTEGRATION.md](docs/AI_AGENT_INTEGRATION.md)（10 分钟）

### 🐍 我想使用 Python 客户端
👉 阅读 [examples/README.md](examples/README.md) 的 Python 部分（5 分钟）

### 📘 我想使用 TypeScript 客户端
👉 阅读 [examples/README.md](examples/README.md) 的 TypeScript 部分（5 分钟）

### 📊 我想了解项目详情
👉 阅读 [COMPLETION_REPORT.md](COMPLETION_REPORT.md)（10 分钟）

---

## 🎯 快速开始（3 步）

### 1️⃣ 构建
```bash
make build
```

### 2️⃣ 配置
```bash
cat > ~/.robotx.yaml << 'YAML'
base_url: https://your-robotx-server.com
api_key: your-api-key
YAML
```

### 3️⃣ 部署
```bash
./robotx deploy ./my-app --name my-app --publish
```

---

## 💡 常见使用场景

### 场景 1: 命令行使用
```bash
# 部署项目
robotx deploy ./my-app --name my-app --publish

# 查询状态
robotx status --project-id proj_xxx

# 查看日志
robotx logs build_xxx
```

### 场景 2: Python 集成
```python
from robotx_client import RobotXClient

client = RobotXClient()
result = client.deploy('./my-app', name='my-app', publish=True)
print(f"Deployed to: {result['url']}")
```

### 场景 3: TypeScript 集成
```typescript
import { RobotXClient } from './robotx_client';

const client = new RobotXClient();
const result = await client.deploy('./my-app', { 
  name: 'my-app', 
  publish: true 
});
console.log(`Deployed to: ${result.url}`);
```

### 场景 4: Claude Desktop
在 Claude Desktop 配置文件中添加：
```json
{
  "mcpServers": {
    "robotx": {
      "command": "/usr/local/bin/robotx",
      "args": ["mcp"]
    }
  }
}
```

---

## 📁 项目文件导航

```
haibingtown/robotx_cli/
│
├── 🚀 快速开始
│   ├── START_HERE.md           ← 你在这里！
│   ├── PROJECT_OVERVIEW.md     ← 项目总览（推荐先读）
│   └── QUICKSTART.md           ← 5 分钟快速入门
│
├── 📖 使用文档
│   ├── README.md               ← 完整使用文档
│   ├── EXAMPLES.md             ← 使用示例集合
│   └── SKILL.md                ← Skill 定义
│
├── 🔧 开发文档
│   ├── docs/
│   │   └── AI_AGENT_INTEGRATION.md  ← AI Agent 集成指南
│   └── examples/
│       ├── README.md           ← 客户端库文档
│       ├── robotx_client.py    ← Python 客户端
│       └── robotx_client.ts    ← TypeScript 客户端
│
├── 📊 项目文档
│   ├── FINAL_SUMMARY.md        ← 最终总结
│   ├── COMPLETION_REPORT.md    ← 完成报告
│   ├── PROJECT_SUMMARY.md      ← 项目总结
│   └── FILES_CREATED.md        ← 文件清单
│
├── 🛠️ 核心代码
│   ├── main.go                 ← 程序入口
│   ├── cmd/                    ← 命令实现
│   └── pkg/client/             ← API 客户端
│
└── 🔨 工具和脚本
    ├── Makefile                ← 构建脚本
    ├── demo.sh                 ← 演示脚本
    └── .robotx.yaml.example    ← 配置示例
```

---

## 🎯 推荐阅读路径

### 路径 1: 快速上手（15 分钟）
1. [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md) - 了解项目
2. [QUICKSTART.md](QUICKSTART.md) - 快速开始
3. 运行 `./demo.sh` - 查看演示

### 路径 2: 深入学习（45 分钟）
1. [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md) - 项目总览
2. [README.md](README.md) - 完整文档
3. [EXAMPLES.md](EXAMPLES.md) - 使用示例
4. [docs/AI_AGENT_INTEGRATION.md](docs/AI_AGENT_INTEGRATION.md) - 集成指南

### 路径 3: AI Agent 开发（30 分钟）
1. [docs/AI_AGENT_INTEGRATION.md](docs/AI_AGENT_INTEGRATION.md) - 集成指南
2. [examples/README.md](examples/README.md) - 客户端库文档
3. 查看 `examples/robotx_client.py` 或 `examples/robotx_client.ts`
4. [EXAMPLES.md](EXAMPLES.md) - 查看更多示例

### 路径 4: 项目管理（20 分钟）
1. [FINAL_SUMMARY.md](FINAL_SUMMARY.md) - 最终总结
2. [COMPLETION_REPORT.md](COMPLETION_REPORT.md) - 完成报告
3. [FILES_CREATED.md](FILES_CREATED.md) - 文件清单

---

## ❓ 常见问题

### Q: 我需要安装什么？
A: 只需要 Go 1.21+ 来构建，或者直接使用编译好的二进制文件。

### Q: 如何配置 RobotX Server 地址？
A: 三种方式：配置文件 `~/.robotx.yaml`、环境变量 `ROBOTX_BASE_URL`、命令行参数 `--base-url`

### Q: 支持哪些编程语言？
A: CLI 本身用 Go 编写，提供 Python 和 TypeScript 客户端库。

### Q: 如何集成到 AI Agent？
A: 查看 [docs/AI_AGENT_INTEGRATION.md](docs/AI_AGENT_INTEGRATION.md)，支持 3 种集成方式。

### Q: 有示例代码吗？
A: 有！查看 [EXAMPLES.md](EXAMPLES.md) 和 `examples/` 目录。

---

## 🆘 需要帮助？

### 📖 查看文档
- [README.md](README.md) - 完整文档
- [QUICKSTART.md](QUICKSTART.md) - 快速入门
- [EXAMPLES.md](EXAMPLES.md) - 使用示例

### 🎬 运行演示
```bash
./demo.sh
```

### 🔍 查看示例
```bash
# Python 示例
cat examples/robotx_client.py

# TypeScript 示例
cat examples/robotx_client.ts
```

---

## ✅ 项目状态

**状态**: ✅ 核心功能已完成，可用于测试和集成  
**版本**: v1.0.0-beta  
**更新**: 2024-02-03

---

## 🎉 开始使用

选择一个适合你的路径，开始探索 RobotX CLI 吧！

**推荐**: 先阅读 [PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)，然后运行 `./demo.sh` 查看演示。

---

**祝你使用愉快！** 🚀
