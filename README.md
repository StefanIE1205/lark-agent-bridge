# lark-agent-bridge

将 Claude Code / Codex / Antigravity CLI 接入 Lark（飞书），在手机上远程指挥本地开发机执行编码任务。

## 工作方式

```
手机 Lark → 飞书长连接 → lab.exe (本机常驻) → Claude/Codex/Antigravity CLI
                                              → 回传进度/结果/错误到 Lark
```

## 前置要求

- **Go 1.23+**（仅构建需要）
- **Lark 企业自建应用**（需管理员权限创建）
- 至少一个本地 Agent CLI：Claude Code、Codex 或 Antigravity

## 快速开始

### 1. 构建

```powershell
git clone https://github.com/StefanIE1205/lark-agent-bridge.git
cd lark-agent-bridge
go build -o lab.exe ./cmd/lab/
```

### 2. 配置 Lark 应用

1. 打开 [飞书开放平台](https://open.feishu.cn/) → 创建企业自建应用
2. **添加能力** → 开启「机器人」
3. **权限管理** → 添加权限：
   - `im:message`（接收消息）
   - `im:message:send_as_bot`（发送消息）
4. **事件订阅** → 订阅 `im.message.receive_v1` → 使用长连接模式
5. **发布** → 创建版本并发布（仅应用管理员可见即可）
6. 复制 `app_id` 和 `app_secret`，以及你的 **user_id**（在「成员管理」中查看）

### 3. 配置文件

```toml
# config.toml
data_dir = "~/.lark-agent-bridge"
default_agent = "codex"

[lark]
app_id = "cli_xxxxxxxx"
app_secret = "your_app_secret"
admin_user_ids = ["ou_xxxxxxxx"]   # 你的 user_id
allowed_chat_ids = []              # 群聊白名单（可选）
require_mention_in_group = true

[[projects]]
name = "demo"
path = "C:\\Users\\me\\my-project"

[[agents]]
name = "claude"
type = "claude"
command = "claude"
args = []

[[agents]]
name = "codex"
type = "codex"
command = "codex"
args = []
```

完整示例参见 `config.example.toml`。

### 4. 启动

```powershell
.\lab.exe --config config.toml
```

看到 `ready — waiting for messages` 即启动成功。

### 5. 手机 Lark 测试

在 Lark 中找到你的机器人，发送：

```
/ping
```

收到 `pong` 表示链路已通。

## 支持的命令

| 命令 | 说明 | 权限 |
|------|------|------|
| `/ping` | 测试连接 | 所有人 |
| `/help` | 查看帮助 | 所有人 |
| `/bind <name> <path>` | 绑定项目目录 | 管理员 |
| `/project <name>` | 设置当前项目 | 所有人 |
| `/agent <name>` | 设置当前 Agent | 所有人 |
| `/ask <task>` | 发送任务 | 管理员 |
| `! <task>` | `/ask` 的简写 | 管理员 |
| `/status` | 查看当前状态 | 所有人 |
| `/sessions` | 查看活跃会话 | 所有人 |
| `/stop` | 停止当前任务 | 管理员 |
| `/log [n]` | 查看最近日志 | 所有人 |
| `/approve <id>` | 批准高风险操作 | 管理员 |
| `/deny <id>` | 拒绝高风险操作 | 管理员 |

## Agent 安装

### Claude Code

```powershell
npm install -g @anthropic-ai/claude-code
claude --version
```

### Codex (OpenAI)

```powershell
npm install -g @openai/codex
codex --version
```

### Antigravity (实验性)

请参考 Antigravity 文档安装 CLI。如果 CLI 无法被 headless 驱动，服务会报清晰错误，不阻塞运行。

## 目录结构

```
~/.lark-agent-bridge/
  state/
    projects.json          # 项目绑定
    chat_defaults.json     # 会话默认值
  logs/
    app.log                # 应用日志
    sessions/              # 会话日志
  audit/
    audit.log              # 审计日志
```

## 开发

```powershell
# 运行所有测试
go test ./...

# 构建
go build -o lab.exe ./cmd/lab/

# 使用 Makefile
make build
make test
```

## 安全

- 所有外部输入视为不可信，严格校验
- 非管理员无法执行 `/bind`、`/ask`、`/stop`、`/approve`、`/deny`
- 群聊必须在白名单中且 @机器人 才会响应
- 日志和回传内容自动脱敏（API_KEY、TOKEN 等）
- 本地命令执行默认关闭（`allow_shell = false`）
- 高风险操作需管理员审批

## 限制（MVP）

- 仅支持 Lark（飞书）单平台
- 不支持多租户 / SaaS
- 不支持复杂卡片 UI
- 不支持 PTY（Windows 用 stdin/stdout pipe）
- Antigravity 标记为实验性

## License

MIT
