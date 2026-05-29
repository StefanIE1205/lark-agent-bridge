# lark-agent-bridge V2 自然语言体验优化规格

本文档基于当前第一版代码审查结果、`cc-connect` 的工程经验，以及你希望达到的体验目标编写。

第二版的核心目标不是“增加更多命令”，而是把 `lark-agent-bridge` 从一个命令机器人升级成一个自然语言远程开发助手：

> 在 Lark 里像和一个开发同事聊天一样说话，它能理解上下文、自动选择项目和 Agent、主动追问缺失信息、持续汇报状态，并在风险操作前请求确认。

---

## 1. 当前第一版审查结论

当前仓库：`C:\Users\22611\my-project`

远程仓库：`https://github.com/StefanIE1205/lark-agent-bridge.git`

当前本地 HEAD：

```text
bfb4725 fix: adopt cc-connect patterns for OpenId priority and bot self-discovery
```

本地未跟踪文件：

```text
start.bat
```

### 1.1 第一版已经完成的基础

第一版已经有清晰的基础骨架：

- `internal/lark`：Lark 长连接、消息解析、回复发送。
- `internal/core`：命令路由、任务入口、进度聚合。
- `internal/agent`：Claude、Codex、Antigravity、Fake Agent adapter。
- `internal/session`：Session 状态机。
- `internal/security`：权限策略、审批请求、脱敏。
- `internal/store`：项目绑定和 chat defaults JSON 持久化。
- `cmd/lab`：可执行入口。

这说明第一版已经完成了“机器人能连上 Lark、能收消息、能触发本地 Agent”的骨架。

### 1.2 第一版不符合预期的根因

第一版的交互模型仍然是命令式：

```text
/project demo
/agent codex
/ask 帮我修复测试失败
```

这会带来几个问题：

- 用户必须记命令。
- 聊天体验不像真实助手。
- Agent 选择、项目选择、会话选择都暴露给用户。
- Lark 群聊里沟通成本高，不像“直接吩咐一个同事”。
- 机器人缺少主动理解和追问能力。

你想要的是：

```text
帮我看看昨天那个登录页报错，能不能直接修掉
```

机器人应该自己判断：

- 这是一个开发任务。
- 应该使用哪个项目。
- 应该用哪个 Agent。
- 需要先检查日志、测试还是 git diff。
- 信息不够时应该追问。
- 风险操作需要确认。

---

## 2. 第二版产品目标

### 2.1 体验目标

第二版应达到以下体验：

1. **默认自然语言入口**
   用户发普通话即可触发任务，不需要 `/ask`。

2. **命令退到后台**
   Slash commands 只作为高级控制手段保留，不作为主交互路径。

3. **上下文记忆**
   机器人记住每个聊天/话题的默认项目、最近任务、最近 Agent、当前分支和上次失败点。

4. **主动追问**
   信息不足时不要直接失败，而是追问：

   ```text
   你说的“后台项目”我这里有 backend-api 和 admin-server 两个绑定项目，要我看哪个？
   ```

5. **自动选择 Agent**
   默认由系统根据任务类型选择 Agent，也允许用户自然语言指定：

   ```text
   这次让 Claude 来做
   用 Codex 帮我跑一遍
   ```

6. **低噪声进度**
   不刷屏。只给“已接收、正在做、卡住了、完成了、需要你确认”这几类关键反馈。

7. **风险前确认**
   安装依赖、删除文件、推送代码、部署、执行任意 shell 前必须确认。

8. **像真实开发助手一样回复**
   最终回复不只是 stdout，而是结构化总结：

   ```text
   修好了。

   我改了：
   - internal/auth/session.go：修复 token 过期时空指针
   - internal/auth/session_test.go：补了回归测试

   验证：
   - go test ./internal/auth 通过

   还没做：
   - 没有提交 commit，等你确认
   ```

### 2.2 第二版非目标

第二版仍然不做：

- SaaS。
- 多平台机器人。
- 完整 Web Admin。
- 插件市场。
- 多人复杂权限系统。
- 复杂低代码工作流。

第二版只做一件事：让 Lark 上的自然语言远程开发体验变顺。

---

## 3. Hermes-agent 式体验拆解

你提到的“像 Hermes-agent 接入飞书一样流畅”，可以拆成下面几类能力。

### 3.1 消息即任务

用户不需要显式输入 `/ask`。机器人应该将大部分普通消息识别成下列意图之一：

- 开发任务：修复、实现、重构、测试、解释代码。
- 查询任务：状态、日志、当前进度、最近修改。
- 控制任务：停止、切换项目、切换 Agent、确认/拒绝。
- 闲聊/无关消息：忽略或轻量回应。

### 3.2 群聊中自然触发

群聊里只要 @ 机器人即可：

```text
@lab 帮我看看这个仓库为什么启动不了
```

不需要：

```text
@lab /ask 帮我看看这个仓库为什么启动不了
```

### 3.3 回复形态要像人

第一版现在偏机器：

```text
已收到任务，正在处理...
会话：xxx
项目：demo
Agent：codex
```

第二版应更自然：

```text
收到。我会在 demo 项目里用 Codex 先跑测试，再看失败点。
```

### 3.4 状态变化要即时但克制

建议状态反馈：

- 收到：立即回复。
- 开始：说明项目和 Agent。
- 进展：最多 15-30 秒一次摘要。
- 等确认：强提醒。
- 完成：结构化报告。

不要逐行转发 Agent 输出。

---

## 4. 第二版架构

第二版新增一层 `Conversation Orchestrator`，把当前“命令路由器”升级成“对话编排器”。

```text
Lark Adapter
  |
  v
Message Normalizer
  |
  v
Conversation Orchestrator
  |
  +--> Intent Router
  +--> Context Resolver
  +--> Project Resolver
  +--> Agent Selector
  +--> Clarification Manager
  +--> Approval Coordinator
  +--> Response Composer
  |
  v
Session Broker
  |
  v
Agent Runtime
```

### 4.1 新增模块

建议新增目录：

```text
internal/conversation/
  orchestrator.go
  intent.go
  context.go
  resolver.go
  clarification.go
  composer.go
  memory.go
  orchestrator_test.go
  intent_test.go
  resolver_test.go

internal/runtime/
  broker.go
  task.go
  result.go
  broker_test.go
```

### 4.2 现有模块调整

`internal/core`

- 保留命令解析。
- 不再让命令系统成为唯一入口。
- 新增 `HandleNaturalMessage` 或将 `HandleMessage` 改成先走 conversation orchestrator。

`internal/session`

- Session 不只保存状态，还要保存“最近任务摘要”和“当前运行目标”。

`internal/store`

- 增加 conversation memory、message dedup 持久化、pending clarification。

`internal/agent`

- 修复 workdir 没有传入的问题。
- 增强为真正的任务运行层，而不是一次性命令 pipe。

---

## 5. 第二版核心数据结构

### 5.1 Intent

```go
type IntentType string

const (
    IntentAsk          IntentType = "ask"
    IntentStatus       IntentType = "status"
    IntentStop         IntentType = "stop"
    IntentSwitchProject IntentType = "switch_project"
    IntentSwitchAgent  IntentType = "switch_agent"
    IntentApprove      IntentType = "approve"
    IntentDeny         IntentType = "deny"
    IntentHelp         IntentType = "help"
    IntentIgnore       IntentType = "ignore"
    IntentClarifyReply IntentType = "clarify_reply"
)

type Intent struct {
    Type       IntentType
    Task       string
    ProjectHint string
    AgentHint   string
    Confidence  float64
    Raw         string
}
```

### 5.2 ConversationContext

```go
type ConversationContext struct {
    ChatID        string
    ThreadID      string
    UserID        string
    IsDirect      bool
    LastProject   string
    LastAgent     string
    LastTask      string
    LastTaskStatus string
    PendingQuestionID string
    UpdatedAt     time.Time
}
```

### 5.3 TaskEnvelope

```go
type TaskEnvelope struct {
    ID          string
    SessionKey  string
    UserText    string
    NormalizedTask string
    Project     string
    WorkDir     string
    Agent       string
    SourceMessageID string
    CreatedAt   time.Time
}
```

### 5.4 Clarification

```go
type Clarification struct {
    ID        string
    ChatID    string
    ThreadID  string
    Question  string
    Options   []string
    OriginalMessage string
    CreatedAt time.Time
    ExpiresAt time.Time
}
```

---

## 6. 自然语言入口规则

### 6.1 入口优先级

收到消息后按这个顺序处理：

1. 安全校验。
2. 消息去重。
3. 如果是 slash command，走 legacy command router。
4. 如果是 pending clarification 的回复，走 clarification resolver。
5. 否则走 intent router。
6. intent 为 ask 时创建任务。
7. intent 不确定时追问。

### 6.2 自然语言示例

用户说：

```text
帮我跑一下测试，看看为什么失败
```

系统行为：

- 如果当前 chat/thread 有默认项目，直接执行。
- 如果没有默认项目，但只绑定了一个项目，自动使用。
- 如果绑定多个项目，追问。

用户说：

```text
用 Claude 帮我看一下这个模块怎么重构
```

系统行为：

- 识别 AgentHint=`claude`。
- 识别任务为“看模块重构”。
- 项目不明确则使用上下文或追问。

用户说：

```text
停一下
```

系统行为：

- intent=`stop`。
- 停止当前 thread/session 任务。

用户说：

```text
可以，继续
```

系统行为：

- 如果存在 pending approval，视为 approve。
- 如果存在 pending clarification，视为选择默认选项。
- 否则作为普通任务上下文追加。

---

## 7. Project Resolver

当前第一版要求用户显式 `/project demo`。第二版改成自动解析。

### 7.1 解析顺序

1. 消息里明确提到项目名：

   ```text
   帮我看 backend 项目
   ```

2. 当前 thread 绑定过项目。
3. 当前 chat 绑定过项目。
4. 只有一个项目时自动使用。
5. 多项目且无法判断时追问。

### 7.2 追问示例

```text
我需要先确认项目。你现在绑定了：

1. lark-agent-bridge
2. my-project
3. backend-api

回复序号或项目名即可。
```

---

## 8. Agent Selector

当前第一版要求用户显式 `/agent codex`。第二版改成自动选择。

### 8.1 默认策略

配置：

```toml
[agent_selection]
default = "codex"
code_review = "claude"
implementation = "codex"
exploration = "claude"
experimental = "antigravity"
```

### 8.2 规则

- “实现、修复、跑测试、改代码”：默认 Codex。
- “解释、架构、重构建议、代码审查”：默认 Claude。
- “探索式、多文件规划、实验”：可尝试 Antigravity。
- 用户自然语言指定时优先用户指定：

  ```text
  这次让 Claude 做
  用 Codex 跑一下
  交给 Antigravity 探索一下
  ```

---

## 9. Response Composer

第二版要避免把 Agent stdout 直接丢给用户。

### 9.1 ACK 模板

```text
收到。我会在 {project} 里用 {agent} 处理这件事。
```

如果需要先确认：

```text
我还差一个信息：你要我看哪个项目？
```

### 9.2 进度模板

```text
还在处理。当前进展：
- 已定位到失败测试
- 正在检查 session 初始化逻辑
```

### 9.3 完成模板

```text
完成了。

我做了：
- ...

验证：
- ...

改动：
- ...

还需要你确认：
- ...
```

### 9.4 失败模板

```text
我卡住了。

原因：
- ...

我已经尝试：
- ...

你可以：
- 回复“继续排查”
- 回复“换 Claude 看看”
- 回复“停止”
```

---

## 10. 第一版必须优先修复的问题

这些不是体验优化，而是第二版前置修复。

### F001 Agent WorkDir 没有传入

当前 `core.Engine.runAgent` 中：

```go
opts := agent.SessionOptions{
    SessionKey: key.String(),
    WorkDir:    "",
}
```

这会导致 Agent 不一定在绑定项目目录中工作。

修复：

- `runAgent` 必须接收 `workDir`。
- `handleAsk` 调用时传入 resolved workDir。
- 测试 fake agent 记录 opts.WorkDir。

### F002 并不是真正的持久会话

当前 Runner 是每次 `Send` 启动进程，写入 prompt 后关闭 stdin。

这更像一次性任务，不像“持续对话”。

第二版要求：

- `AgentSession` 应该能跨多轮复用。
- 如果某个 Agent 暂时不支持持久进程，也要显式标记 `SessionMode=one-shot`。
- 会话记忆由 conversation 层补齐，不要假装 CLI session 已经持久。

### F003 `/approve` 和 `/deny` 尚未真正接入审批系统

当前 `handleApproval` 只是回复“已记录”，没有调用 `ApprovalManager`。

第二版要求：

- ApprovalManager 注入 Engine。
- approve/deny 改变 pending request 状态。
- 如果 Agent 正在等待，能恢复或终止任务。

### F004 群聊 mention 判断过宽

当前 `isMentioned` 只要有任何 mention 就返回 true。

第二版要求：

- 只在 mention 目标是当前 bot 时返回 true。
- 或在 parse 阶段根据 botOpenID 精准匹配。

### F005 空 admin 列表允许所有人

当前 `Policy.IsAdmin` 中空 admin list 等于 allow all。

第二版建议：

- 本地开发可通过 `dev_allow_all = true` 显式打开。
- 生产默认必须至少一个 admin。

### F006 进度结束时可能重复发送

当前 `ProgressReporter.Final()` 会调用 `sendFn`，随后 `runAgent` 又发送“任务完成”。

第二版要求：

- 最终回复只发送一次。
- 中间进度和最终报告分离。

### F007 `progress_interval_ms` 配置未使用

当前 `progressInterval()` 固定返回 3 秒。

第二版要求：

- EngineConfig 带入 progress interval。
- 测试覆盖配置值。

### F008 chat defaults 只按 chat 维度保存

当前 `SetChatDefault(chatID, ...)` 没有 thread 维度。

第二版要求：

- 私聊可按 chat 保存。
- 群聊必须按 `chat_id + thread_id` 保存上下文。

---

## 11. 第二版任务清单

### P0：前置修复

#### V2-T001 修复 WorkDir 传递

目标：

- 确保 Agent 在绑定项目目录中运行。

验收：

- fake agent 能断言收到正确 WorkDir。
- Claude/Codex adapter 的 RunnerConfig.WorkDir 不为空。

#### V2-T002 接入 progress_interval_ms

目标：

- 使用配置中的进度间隔。

验收：

- 配置 `progress_interval_ms = 15000` 时，进度不再 3 秒刷一次。

#### V2-T003 精准 bot mention 判断

目标：

- 群聊只响应 @ 当前 bot 的消息。

验收：

- @ 其他人时不响应。
- @ 当前 bot 时响应。

#### V2-T004 收紧 admin 默认策略

目标：

- 空 admin 不再默认允许所有人。

验收：

- 未配置 admin 且未开启 dev_allow_all 时启动失败。

### P1：自然语言入口

#### V2-T101 Intent Router

新增：

```text
internal/conversation/intent.go
```

目标：

- 将普通文本识别成 ask/status/stop/switch_project/switch_agent/approve/deny/help/ignore。

第一版规则引擎即可，不需要 LLM：

- 包含“停止、别做了、停一下” => stop。
- 包含“状态、进度、做到哪” => status。
- 包含“用 Claude / 让 Claude” => AgentHint=claude。
- 包含“用 Codex / 让 Codex” => AgentHint=codex。
- 包含“帮我、看一下、修复、实现、跑测试、解释、总结” => ask。

验收：

- `帮我跑一下测试` => ask。
- `停一下` => stop。
- `现在做到哪了` => status。
- `用 Claude 看一下这个模块` => ask + AgentHint=claude。

#### V2-T102 Message Entry 改造

目标：

- `Engine.HandleMessage` 先判断 slash command；非 slash 文本进入 Conversation Orchestrator。

验收：

- 不带 `/ask` 的普通消息也能启动任务。
- 原有 `/ping`、`/status` 等命令仍可用。

#### V2-T103 自然语言 ACK

目标：

- 把“已收到任务，正在处理...”改成自然确认。

验收：

```text
帮我跑一下测试
```

回复类似：

```text
收到。我会在 lark-agent-bridge 里用 Codex 跑测试并总结结果。
```

### P2：上下文和追问

#### V2-T201 Conversation Memory

新增：

```text
internal/conversation/memory.go
```

目标：

- 持久化 chat/thread 的默认项目、默认 Agent、最近任务。

存储：

```text
state/conversations.json
```

验收：

- 群聊不同 thread 可以有不同项目上下文。
- 重启后上下文仍存在。

#### V2-T202 Project Resolver

目标：

- 根据消息、thread memory、chat memory、项目数量自动选择项目。

验收：

- 只有一个项目时普通消息直接执行。
- 多项目无上下文时追问。
- 消息里提到项目名时自动选择。

#### V2-T203 Clarification Manager

目标：

- 信息不足时保存 pending question。
- 用户下一条回复可作为答案。

验收：

机器人问：

```text
你要我看哪个项目？1. demo 2. backend
```

用户回复：

```text
2
```

系统继续执行 backend 任务。

#### V2-T204 Agent Selector

目标：

- 根据任务类型和自然语言提示选择 Agent。

验收：

- “解释一下架构”默认 Claude。
- “修一下测试”默认 Codex。
- “用 Claude 修一下”使用 Claude。

### P3：更顺滑的任务执行

#### V2-T301 Session Broker

新增：

```text
internal/runtime/broker.go
```

目标：

- 管理 task queue、session reuse、one-shot/persistent 两种 Agent 模式。

验收：

- 同一个 thread 的连续消息进入同一 session context。
- 忙碌时回复“我还在处理上一件事，要排队还是停止上一件？”

#### V2-T302 Agent Runtime Capability

目标：

- 每个 Agent 暴露能力。

```go
type Capability struct {
    PersistentSession bool
    SupportsApproval  bool
    SupportsStreaming bool
    Experimental      bool
}
```

验收：

- Antigravity experimental 会在 ACK 中轻提示。
- one-shot Agent 不被当作真实持久 session。

#### V2-T303 输出摘要器

目标：

- 不直接转发 stdout，而是整理为进度摘要和最终报告。

MVP 规则：

- 收集输出。
- 提取最后 2000 字。
- 识别测试通过/失败关键词。
- 最终报告分为“结果、改动、验证、下一步”。

验收：

- 最终消息不再是纯 stdout。

### P4：Lark 体验增强

#### V2-T401 Markdown/Post 回复

目标：

- 长结果使用 Lark post/markdown 格式，而不是纯文本。

验收：

- 文件列表、测试结果、代码块可读性明显提升。

#### V2-T402 消息更新或状态消息合并

目标：

- 优先更新同一条状态消息，减少刷屏。

验收：

- 一次任务最多产生：ACK、状态更新、最终报告、审批请求。

#### V2-T403 审批文本升级

目标：

- 第一版仍可用 `/approve id`，但自然语言也能批准。

验收：

- “可以，继续” => approve。
- “别执行这个” => deny。

#### V2-T404 附件和图片入口

目标：

- 用户发截图/日志文件时，作为任务上下文传给 Agent 或保存并提示。

验收：

- Lark 图片消息不会被忽略。
- 系统能回复“我收到了截图，会作为上下文一起分析”。

---

## 12. 第二版配置建议

```toml
data_dir = "~/.lark-agent-bridge"
default_agent = "codex"
default_project = "lark-agent-bridge"
progress_interval_ms = 15000

[lark]
app_id = "cli_xxx"
app_secret = "xxx"
admin_user_ids = ["ou_xxx"]
allowed_chat_ids = []
require_mention_in_group = true

[conversation]
natural_language_enabled = true
auto_project_resolution = true
auto_agent_selection = true
clarification_timeout_secs = 600
reply_style = "concise"

[agent_selection]
default = "codex"
implementation = "codex"
debugging = "codex"
review = "claude"
architecture = "claude"
experimental = "antigravity"

[security]
dev_allow_all = false
allow_shell = false
require_confirm_for_write = true
require_confirm_for_install = true
require_confirm_for_git_push = true
redact_secrets = true
```

---

## 13. 第二版验收标准

### 13.1 自然语言主链路

用户在 Lark 发：

```text
帮我跑一下这个项目的测试，看看有没有失败
```

系统应：

1. 自动识别为开发任务。
2. 自动选择项目。
3. 自动选择 Codex。
4. 回复自然 ACK。
5. 执行任务。
6. 低频汇报进度。
7. 返回结构化结果。

### 13.2 多项目追问链路

用户发：

```text
帮我看一下启动失败
```

如果没有上下文且有多个项目，系统应追问：

```text
你要我看哪个项目？回复序号即可。
```

用户回复：

```text
1
```

系统继续执行。

### 13.3 Agent 自然切换

用户发：

```text
这次让 Claude 帮我审一下这个模块
```

系统应使用 Claude，不要求 `/agent claude`。

### 13.4 控制自然语言

用户发：

```text
停一下
```

系统应停止当前任务，不要求 `/stop`。

### 13.5 审批自然语言

系统请求确认：

```text
我需要执行 npm install，允许吗？
```

用户回复：

```text
可以
```

系统应批准，而不是要求 `/approve id`。

---

## 14. 推荐开发顺序

严格建议：

1. 先修 `F001-F008`。
2. 再做 `Intent Router`。
3. 再改 `HandleMessage` 入口。
4. 再做 project/agent resolver。
5. 再做 clarification。
6. 最后做 Lark post/markdown 和消息更新。

不要先做漂亮卡片。第二版的核心不是 UI，而是“不用命令也能顺畅交流”。

---

## 15. 成功标准

第二版成功时，你应该可以在手机 Lark 中这样连续交流：

```text
帮我看看这个项目为什么启动不了
```

机器人：

```text
收到。我会在 lark-agent-bridge 里用 Codex 先检查启动脚本和最近日志。
```

你：

```text
如果是配置问题就直接修
```

机器人：

```text
明白。我会优先修配置类问题；如果涉及安装依赖或删除文件，会先问你。
```

机器人稍后：

```text
找到原因了：启动时没有把工作目录传给 Agent，导致它不在绑定项目里执行。

我建议修 internal/core/engine.go，并补一个 fake agent 测试。要我直接改吗？
```

你：

```text
改
```

这就是第二版要达到的感觉。

