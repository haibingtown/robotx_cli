# Start Here

如果你第一次使用 `robotx_cli`，按下面顺序阅读：

1. [README.md](README.md): 当前可用功能、命令契约、Action/Release 说明
2. [QUICKSTART.md](QUICKSTART.md): 最短可执行流程
3. [docs/AI_AGENT_INTEGRATION.md](docs/AI_AGENT_INTEGRATION.md): Agent/CI 集成方式
4. [SKILL.md](SKILL.md): Skill 使用说明

## 最短路径

```bash
curl -fsSL https://raw.githubusercontent.com/haibingtown/robotx_cli/main/scripts/install.sh | bash
export ROBOTX_BASE_URL=https://your-robotx-server.com
export ROBOTX_API_KEY=your-api-key
robotx deploy . --name my-app --output json
```

## 当前限制

- `robotx mcp` 还未实现，不建议用于生产集成
- 推荐使用 shell/CLI + `--output json` 作为稳定接口
