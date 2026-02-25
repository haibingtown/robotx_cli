# RobotX CLI

RobotX CLI 用于将应用部署到 RobotX 平台，支持 `deploy` / `update` / `status` / `logs` / `publish`。

## 当前状态

- CLI 集成（shell/CI/Agent）: 可用
- JSON 机器输出: 可用（`--output json` 或 `--json`）
- MCP 模式（`robotx mcp`）: 未实现（占位）

## 安装

### 方式 1: 下载安装脚本（推荐，无需 Go）

```bash
curl -fsSL https://raw.githubusercontent.com/haibingtown/robotx_cli/main/scripts/install.sh | bash
```

可选参数：

- `ROBOTX_VERSION=latest`（默认）或 `vX.Y.Z`
- `ROBOTX_INSTALL_DIR=$HOME/.local/bin`
- `ROBOTX_REPO=haibingtown/robotx_cli`
- `ROBOTX_AUTO_PATH=1`（默认，自动写入 shell profile）或 `0`

### 方式 2: 从源码安装

```bash
go install github.com/haibingtown/robotx_cli/cmd/robotx@latest
```

### 方式 3: 使用 Go 安装并自动配置 PATH

```bash
curl -fsSL https://raw.githubusercontent.com/haibingtown/robotx_cli/main/scripts/go-install.sh | bash
```

可选参数：

- `ROBOTX_GO_PACKAGE=github.com/haibingtown/robotx_cli/cmd/robotx@latest`
- `ROBOTX_LEGACY_GO_PACKAGE=github.com/haibingtown/robotx_cli@latest`（主包安装失败时回退）
- `ROBOTX_INSTALL_DIR=$HOME/.local/bin`
- `ROBOTX_AUTO_PATH=1`（默认，自动写入 shell profile）或 `0`

说明：纯 `go install ...` 命令本身不会自动修改你的 shell 环境变量（PATH），这是 Go 工具链行为；如需“安装后直接可用”建议用方式 1 或方式 3。

## 配置

支持配置文件 `~/.robotx.yaml`：

```yaml
base_url: https://api.robotx.xin
api_key: your-api-key
```

或使用环境变量：

```bash
export ROBOTX_BASE_URL=https://api.robotx.xin
export ROBOTX_API_KEY=your-api-key
```

## 输出模式

- `--output text`（默认）: 面向人类阅读
- `--output json` 或 `--json`: 面向程序解析

在 JSON 模式下：

- stdout: 仅 JSON 结果
- stderr: 进度日志/诊断信息

成功输出结构：

```json
{
  "success": true,
  "command": "deploy",
  "data": {
    "project_id": "proj_xxx",
    "build_id": "build_xxx"
  }
}
```

失败输出结构（stderr 最后一行）：

```json
{
  "success": false,
  "error": {
    "code": "api_error",
    "message": "failed to create project"
  }
}
```

## 命令

### deploy

部署新项目或已有项目。

```bash
robotx deploy [project-path] \
  [--name my-app | --project-id proj_123] \
  [--publish] [--wait=true] [--timeout 600]
```

本地构建模式：

```bash
robotx deploy . --name my-app --local-build \
  [--install-command "npm ci"] \
  [--build-command "npm run build"] \
  [--output-dir dist]
```

### update

更新已有项目（本质复用 deploy 流程）：

```bash
robotx update [project-path] --project-id proj_123 [--publish]
```

### status

查询项目和/或构建状态：

```bash
robotx status [--project-id proj_123] [--build-id build_456] [--logs]
```

说明：

- `--project-id` 与 `--build-id` 至少提供一个
- 指定 `--logs` 时必须提供 `--build-id`

### logs

独立日志查询命令：

```bash
robotx logs --build-id build_456 [--project-id proj_123]
```

### publish

发布构建到生产环境：

```bash
robotx publish --project-id proj_123 --build-id build_456
```

### mcp

```bash
robotx mcp
```

当前返回未实现错误（占位功能）。

## GitHub Action

仓库根目录提供了 composite action（[action.yml](action.yml)），默认流程是：

1. 下载 release 二进制
2. 校验 checksum
3. 执行 `robotx deploy --output json`
4. 输出 `project_id/build_id/status/url` 等字段

示例工作流见：`.github/workflows/action-example.yml`。

补充：

- 支持输入别名：`base_url`/`api_key`（等价于 `base-url`/`api-key`）
- `version: source` 可在 CI 中直接从 action 源码构建 CLI（适合验证 `@main` 最新变更）

## Release

标签推送触发自动发布：

- Workflow: `.github/workflows/release.yml`
- 产物：
  - `robotx_<version>_<os>_<arch>.tar.gz`（linux/darwin）
  - `robotx_<version>_<os>_<arch>.zip`（windows）
  - `checksums.txt`

## 退出码

- `1`: 参数/配置/通用错误
- `2`: API/网络错误
- `3`: 构建失败
- `4`: 发布失败
