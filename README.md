# lark-agent-bridge

将 Claude Code / Codex / Antigravity CLI 接入 Lark（飞书），在手机上远程指挥本地开发机执行编码任务。

## 工作方式

```
手机 Lark → 飞书长连接 → lab.exe (本机常驻) → Claude/Codex/Antigravity CLI
                                              → 回传进度/结果/错误到 Lark
```

## 特性

- **自然语言交互**：直接说"帮我跑测试"，不需要 `/ask` 前缀
- **智能意图识别**：自动识别 ask/stop/status/help/approve/deny
- **上下文记忆**：记住每个对话的项目和 Agent 设置
- **自动选择**：根据任务类型自动选择 Agent（review→claude, implementation→codex）
- **低噪声**：优先更新同一条消息，减少刷屏
- **安全审批**：高风险操作需确认，支持自然语言审批

---

## 前置要求

### 软件要求

| 软件 | 版本 | 用途 | 安装方式 |
|------|------|------|----------|
| **Go** | 1.23+ | 构建项目 | [go.dev/dl](https://go.dev/dl/) |
| **Node.js** | 18+ | 安装 Agent CLI | [nodejs.org](https://nodejs.org/) |
| **Git** | 任意 | 克隆仓库 | [git-scm.com](https://git-scm.com/) |

### Agent CLI 要求（至少安装一个）

| Agent | 安装命令 | 验证命令 |
|-------|----------|----------|
| **Claude Code** | `npm install -g @anthropic-ai/claude-code` | `claude --version` |
| **Codex** | `npm install -g @openai/codex` | `codex --version` |
| **Antigravity** | 参考 [Antigravity 文档](https://antigravity.dev) | `agy --version` |

### API 和 Secret 要求

| 项目 | 来源 | 说明 |
|------|------|------|
| **Lark App ID** | 飞书开放平台 | 企业自建应用的 App ID |
| **Lark App Secret** | 飞书开放平台 | 企业自建应用的 App Secret |
| **Claude API Key** | Anthropic Console | Claude Code 需要（`claude login`） |
| **OpenAI API Key** | OpenAI Platform | Codex 需要（`codex login`） |

---

## 详细部署教程

### 第一步：创建 Lark 企业自建应用

1. **打开飞书开放平台**
   - 国际版：https://open.larksuite.com/
   - 中国版：https://open.feishu.cn/

2. **创建应用**
   - 点击「创建企业自建应用」
   - 填写应用名称（如"我的开发助手"）
   - 选择应用图标

3. **添加机器人能力**
   - 进入应用 → 「添加能力」→ 开启「机器人」

4. **配置权限**
   - 进入「权限管理」→ 搜索并添加：
     - `im:message` — 接收消息
     - `im:message:send_as_bot` — 以机器人身份发送消息
     - `im:resource` — 获取消息中的资源（图片/文件）

5. **配置事件订阅**
   - 进入「事件订阅」→ 选择「使用长连接模式」
   - 添加事件：`im.message.receive_v1`（接收消息）

6. **发布应用**
   - 进入「版本管理与发布」→ 创建版本
   - 设置可用范围（建议先选"仅应用管理员"测试）
   - 发布并等待审核通过

7. **记录关键信息**
   - **App ID**：在「凭证与基础信息」中复制
   - **App Secret**：在「凭证与基础信息」中复制
   - **你的 User ID**：在「成员管理」中查看你的 `open_id`（格式：`ou_xxxxxxxx`）

### 第二步：安装 Agent CLI

#### 安装 Claude Code

```bash
# 安装
npm install -g @anthropic-ai/claude-code

# 验证
claude --version

# 登录（需要 Anthropic API Key）
claude login
```

#### 安装 Codex（可选）

```bash
# 安装
npm install -g @openai/codex

# 验证
codex --version

# 登录（需要 OpenAI API Key）
codex login
```

#### 安装 Antigravity（可选，实验性）

参考 [Antigravity 官方文档](https://antigravity.dev) 安装 CLI。

### 第三步：构建项目

```bash
# 克隆仓库
git clone https://github.com/StefanIE1205/lark-agent-bridge.git
cd lark-agent-bridge

# 设置 Go 代理（中国大陆）
go env -w GOPROXY=https://goproxy.cn,direct

# 构建
go build -o lab.exe ./cmd/lab/

# 验证
.\lab.exe --version
```

### 第四步：配置

```bash
# 复制示例配置
cp config.example.toml config.toml

# 编辑配置
# Windows: notepad config.toml
# macOS/Linux: nano config.toml
```

**配置文件说明：**

```toml
# 数据目录（存储会话状态、日志等）
data_dir = "~/.lark-agent-bridge"

# 默认 Agent（claude/codex/antigravity）
default_agent = "claude"

# 进度更新间隔（毫秒）
progress_interval_ms = 3000

[lark]
# 飞书应用凭证
app_id = "cli_xxxxxxxxxxxxxxxx"        # 替换为你的 App ID
app_secret = "xxxxxxxxxxxxxxxxxxxxxxxx" # 替换为你的 App Secret

# 管理员用户 ID（可以有多个）
admin_user_ids = ["ou_xxxxxxxxxxxxxxxx"] # 替换为你的 open_id

# 群聊白名单（留空 = 允许所有群聊）
allowed_chat_ids = []

# 群聊中是否需要 @机器人 才响应
require_mention_in_group = false

# 飞书域名（国际版用 larksuite.com，中国版用 feishu.cn）
domain = "https://open.larksuite.com"

# 项目绑定（可以有多个）
[[projects]]
name = "my-project"                    # 项目名称
path = "C:\\Users\\me\\my-project"     # 项目绝对路径

# Agent 配置（可以有多个）
[[agents]]
name = "claude"
type = "claude"
command = "claude"                      # CLI 命令
args = ["-p", "--bare", "--output-format", "text"] # Claude Code 推荐参数

[[agents]]
name = "codex"
type = "codex"
command = "codex"
args = []

[[agents]]
name = "antigravity"
type = "antigravity"
command = "agy"
args = []
experimental = true                     # 标记为实验性

[security]
# 开发模式（允许所有用户，仅用于本地开发）
dev_allow_all = false
```

### 第五步：启动服务

```bash
# Windows
.\lab.exe --config config.toml

# macOS/Linux
./lab --config config.toml
```

看到以下输出表示启动成功：

```
[lab] lab 0.1.0 starting
[lab] security: 1 admin(s), 0 allowed chat(s), require_mention=false
[lab] agents: [claude]
[lab] lark adapter starting...
[lab] ready — waiting for messages
```

### 第六步：测试

1. **打开 Lark**，找到你创建的机器人
2. **发送测试消息**：

```
/ping
```

收到 `pong` 表示链路已通。

3. **测试自然语言**：

```
帮我跑一下测试
```

应该收到类似回复：

```
收到。我会在 my-project 里用 claude 处理这件事。
```

---

## 使用指南

### 自然语言交互（推荐）

直接用自然语言和机器人对话：

```
帮我看看这个项目为什么启动不了
```

```
用 Claude 帮我审一下代码
```

```
修复登录页面的 bug
```

```
停一下
```

```
现在什么状态
```

### Slash 命令（高级控制）

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

### 审批流程

当 Agent 需要执行高风险操作（如安装依赖、删除文件）时：

1. 机器人会发送审批请求
2. 你可以回复：
   - `可以` / `继续` / `approve` — 批准
   - `不行` / `拒绝` / `deny` — 拒绝
   - 或使用命令：`/approve <id>` / `/deny <id>`

---

## 目录结构

```
~/.lark-agent-bridge/
  state/
    projects.json          # 项目绑定
    chat_defaults.json     # 会话默认值
    conversations.json     # 对话上下文
  logs/
    app.log                # 应用日志
    sessions/              # 会话日志
  audit/
    audit.log              # 审计日志
```

---

## 开发

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/core/ -v

# 构建
go build -o lab.exe ./cmd/lab/

# 使用 Makefile
make build
make test
make release
```

---

## 安全

- **输入校验**：所有外部输入视为不可信，严格校验
- **权限控制**：非管理员无法执行 `/bind`、`/ask`、`/stop`、`/approve`、`/deny`
- **群聊白名单**：群聊必须在白名单中且 @机器人 才会响应
- **数据脱敏**：日志和回传内容自动脱敏（API_KEY、TOKEN 等）
- **命令隔离**：本地命令执行默认关闭（`allow_shell = false`）
- **审批机制**：高风险操作需管理员审批

---

## 故障排除

### 机器人不响应

1. 检查服务是否在运行（看终端输出）
2. 检查 Lark 应用是否已发布
3. 检查事件订阅是否配置正确
4. 检查 `admin_user_ids` 是否包含你的 `open_id`

### Claude Code 报错

```bash
# 检查是否安装
claude --version

# 检查是否登录
claude whoami

# 重新登录
claude login
```

### 连接超时

- 检查网络是否可以访问飞书 API
- 国际版：`open.larksuite.com`
- 中国版：`open.feishu.cn`

---

## 限制

- 仅支持 Lark（飞书）单平台
- 不支持多租户 / SaaS
- 不支持复杂卡片 UI
- 不支持 PTY（Windows 用 stdin/stdout pipe）
- Antigravity 标记为实验性

---

## License

MIT
