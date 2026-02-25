# RobotX CLI 快速开始

## 1) 安装（推荐二进制）

```bash
curl -fsSL https://raw.githubusercontent.com/haibingtown/robotx_cli/main/scripts/install.sh | bash
robotx --version
```

## 2) 配置

```bash
export ROBOTX_BASE_URL=https://your-robotx-server.com
export ROBOTX_API_KEY=your-api-key
```

或写入 `~/.robotx.yaml`：

```yaml
base_url: https://your-robotx-server.com
api_key: your-api-key
```

## 3) 部署

```bash
robotx deploy . --name my-app --output json
```

## 4) 查询状态与日志

```bash
robotx status --project-id proj_123 --output json
robotx status --build-id build_456 --output json
robotx logs --build-id build_456 --output json
```

## 5) 发布

```bash
robotx publish --project-id proj_123 --build-id build_456 --output json
```

## 6) 常见参数

- `--output json` / `--json`: 机器可读输出
- `--publish`: 构建成功后自动发布
- `--wait=false`: 不等待构建结束
- `--timeout 900`: 自定义等待超时

## 注意

- `robotx mcp` 当前未实现（占位功能）
- JSON 模式下 stdout 仅输出 JSON，进度日志写入 stderr
