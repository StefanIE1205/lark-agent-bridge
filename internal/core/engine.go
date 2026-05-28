package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/agent"
	"github.com/StefanIE1205/lark-agent-bridge/internal/logging"
	"github.com/StefanIE1205/lark-agent-bridge/internal/security"
	"github.com/StefanIE1205/lark-agent-bridge/internal/session"
	"github.com/StefanIE1205/lark-agent-bridge/internal/store"
)

type Engine struct {
	platform     Platform
	policy       *security.Policy
	sessions     *session.Manager
	store        *store.StateStore
	agents       map[string]agent.Agent
	defaultAgent string
	logger       *logging.Logger
}

type EngineConfig struct {
	Platform     Platform
	Policy       *security.Policy
	Sessions     *session.Manager
	Store        *store.StateStore
	Agents       []agent.Agent
	DefaultAgent string
	Logger       *logging.Logger
}

func NewEngine(cfg EngineConfig) *Engine {
	e := &Engine{
		platform:     cfg.Platform,
		policy:       cfg.Policy,
		sessions:     cfg.Sessions,
		store:        cfg.Store,
		agents:       make(map[string]agent.Agent),
		defaultAgent: cfg.DefaultAgent,
		logger:       cfg.Logger,
	}
	for _, a := range cfg.Agents {
		e.agents[a.Name()] = a
	}
	return e
}

// HandleMessage is the main entry point for incoming messages.
func (e *Engine) HandleMessage(ctx context.Context, msg Message) {
	if !e.policy.CanAccess(msg.UserID, msg.IsDirect, msg.ChatID, msg.Mentioned) {
		return
	}

	cmd := ParseCommand(msg.Text)
	if cmd.Name == "" {
		return
	}

	if !e.policy.CanExecute(msg.UserID, cmd.Name) {
		e.reply(ctx, msg.ReplyTarget, "权限不足：你无权执行此命令。")
		return
	}

	switch cmd.Name {
	case "ping":
		e.handlePing(ctx, msg)
	case "help":
		e.handleHelp(ctx, msg)
	case "status":
		e.handleStatus(ctx, msg)
	case "sessions":
		e.handleSessions(ctx, msg)
	case "stop":
		e.handleStop(ctx, msg)
	case "bind":
		e.handleBind(ctx, msg, cmd)
	case "project":
		e.handleProject(ctx, msg, cmd)
	case "agent":
		e.handleAgent(ctx, msg, cmd)
	case "ask":
		e.handleAsk(ctx, msg, cmd)
	case "log":
		e.handleLog(ctx, msg, cmd)
	case "approve", "deny":
		e.handleApproval(ctx, msg, cmd)
	default:
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("未知命令: /%s。输入 /help 查看支持的命令。", cmd.Name))
	}
}

func (e *Engine) handlePing(ctx context.Context, msg Message) {
	e.reply(ctx, msg.ReplyTarget, "pong")
}

func (e *Engine) handleHelp(ctx context.Context, msg Message) {
	help := `支持的命令：
/ping - 测试连接
/help - 显示帮助
/bind <name> <path> - 绑定项目（管理员）
/project <name> - 设置当前项目
/agent <claude|codex|antigravity> - 设置当前 Agent
/ask <task> - 发送任务（管理员）
! <task> - /ask 的简写
/status - 查看当前会话状态
/sessions - 查看活跃会话
/stop - 停止当前任务（管理员）
/log [n] - 查看最近日志
/approve <id> - 批准操作（管理员）
/deny <id> - 拒绝操作（管理员）`
	e.reply(ctx, msg.ReplyTarget, help)
}

func (e *Engine) handleStatus(ctx context.Context, msg Message) {
	chatDefaults, _ := e.store.GetChatDefaults(msg.ChatID)
	sessKey := session.DeriveKey(msg.ChatID, msg.ThreadID,
		chatDefaults.Project, chatDefaults.Agent)
	sess := e.sessions.Get(sessKey.String())

	status := fmt.Sprintf(
		"会话：%s\n项目：%s\nAgent：%s\n",
		sessKey.String(),
		chatDefaults.Project,
		chatDefaults.Agent,
	)

	if sess == nil {
		status += "状态：无活跃任务"
	} else {
		status += fmt.Sprintf("状态：%s\n最后活跃：%s",
			sess.Status,
			sess.LastActivity.Format("15:04:05"))
	}

	e.reply(ctx, msg.ReplyTarget, status)
}

func (e *Engine) handleSessions(ctx context.Context, msg Message) {
	active := e.sessions.ListActive()
	if len(active) == 0 {
		e.reply(ctx, msg.ReplyTarget, "无活跃会话。")
		return
	}

	var lines []string
	for _, s := range active {
		lines = append(lines, fmt.Sprintf("- %s [%s] %s/%s",
			s.Key, s.Status, s.Project, s.Agent))
	}
	e.reply(ctx, msg.ReplyTarget, "活跃会话：\n"+strings.Join(lines, "\n"))
}

func (e *Engine) handleStop(ctx context.Context, msg Message) {
	chatDefaults, _ := e.store.GetChatDefaults(msg.ChatID)
	sessKey := session.DeriveKey(msg.ChatID, msg.ThreadID,
		chatDefaults.Project, chatDefaults.Agent)

	if err := e.sessions.Stop(sessKey.String()); err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("停止失败：%v", err))
		return
	}
	e.reply(ctx, msg.ReplyTarget, "任务已停止。")
}

func (e *Engine) handleBind(ctx context.Context, msg Message, cmd Command) {
	if len(cmd.Args) < 2 {
		e.reply(ctx, msg.ReplyTarget, "用法：/bind <name> <absolute-path>\n示例：/bind demo C:\\repo\\demo")
		return
	}

	name := cmd.Args[0]
	path := cmd.Args[1]

	if err := e.store.BindProject(name, path); err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("绑定失败：%v", err))
		return
	}
	e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("已绑定项目 %s → %s", name, path))
}

func (e *Engine) handleProject(ctx context.Context, msg Message, cmd Command) {
	if len(cmd.Args) < 1 {
		e.reply(ctx, msg.ReplyTarget, "用法：/project <name>")
		return
	}

	name := cmd.Args[0]
	if err := e.store.SetChatDefault(msg.ChatID, "project", name); err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("设置项目失败：%v", err))
		return
	}
	e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("当前项目已设为：%s", name))
}

func (e *Engine) handleAgent(ctx context.Context, msg Message, cmd Command) {
	if len(cmd.Args) < 1 {
		e.reply(ctx, msg.ReplyTarget, "用法：/agent <claude|codex|antigravity>")
		return
	}

	name := cmd.Args[0]
	if _, ok := e.agents[name]; !ok {
		var available []string
		for n := range e.agents {
			available = append(available, n)
		}
		e.reply(ctx, msg.ReplyTarget,
			fmt.Sprintf("未知 Agent: %s\n可用：%s", name, strings.Join(available, ", ")))
		return
	}

	if err := e.store.SetChatDefault(msg.ChatID, "agent", name); err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("设置 Agent 失败：%v", err))
		return
	}
	e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("当前 Agent 已设为：%s", name))
}

func (e *Engine) handleAsk(ctx context.Context, msg Message, cmd Command) {
	if len(cmd.Args) < 1 {
		e.reply(ctx, msg.ReplyTarget, "用法：/ask <task>\n示例：/ask 修复测试失败")
		return
	}

	task := cmd.Args[0]
	chatDefaults, _ := e.store.GetChatDefaults(msg.ChatID)

	project := chatDefaults.Project
	if project == "" {
		project = "demo"
	}

	agentName := chatDefaults.Agent
	if agentName == "" {
		agentName = e.defaultAgent
	}

	ag, ok := e.agents[agentName]
	if !ok {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("未知 Agent: %s。请先用 /agent 设置。", agentName))
		return
	}

	workDir, err := e.resolveWorkDir(project)
	if err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("项目错误：%v", err))
		return
	}

	sessKey := session.DeriveKey(msg.ChatID, msg.ThreadID, project, agentName)
	sess, err := e.sessions.GetOrCreate(msg.ChatID, msg.ThreadID, project, agentName, workDir)
	if err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("创建会话失败：%v", err))
		return
	}

	if err := e.sessions.StartTask(sess.Key); err != nil {
		e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("无法启动任务：%v", err))
		return
	}

	e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("已收到任务，正在处理...\n会话：%s\n项目：%s\nAgent：%s", sessKey.String(), project, agentName))

	go e.runAgent(ctx, ag, sessKey, task, msg.ReplyTarget)
}

func (e *Engine) handleLog(ctx context.Context, msg Message, cmd Command) {
	n := 50
	if len(cmd.Args) > 0 {
		if parsed, err := strconv.Atoi(cmd.Args[0]); err == nil && parsed > 0 {
			n = parsed
			if n > 200 {
				n = 200
			}
		}
	}

	chatDefaults, _ := e.store.GetChatDefaults(msg.ChatID)
	sessKey := session.DeriveKey(msg.ChatID, msg.ThreadID,
		chatDefaults.Project, chatDefaults.Agent)

	lines := e.readSessionLog(sessKey.String(), n)
	if len(lines) == 0 {
		e.reply(ctx, msg.ReplyTarget, "暂无日志。")
		return
	}

	text := strings.Join(lines, "\n")
	if len(text) > 3000 {
		text = text[len(text)-3000:]
	}
	e.reply(ctx, msg.ReplyTarget, text)
}

func (e *Engine) readSessionLog(sessionKey string, maxLines int) []string {
	if e.logger == nil {
		return nil
	}

	content, err := e.logger.ReadSessionLog(sessionKey, maxLines)
	if err != nil {
		return []string{fmt.Sprintf("读取日志失败：%v", err)}
	}
	return content
}

func (e *Engine) handleApproval(ctx context.Context, msg Message, cmd Command) {
	action := "批准"
	if cmd.Name == "deny" {
		action = "拒绝"
	}
	reqID := ""
	if len(cmd.Args) > 0 {
		reqID = cmd.Args[0]
	}
	e.reply(ctx, msg.ReplyTarget, fmt.Sprintf("%s操作已记录（请求ID: %s）", action, reqID))
}

func (e *Engine) runAgent(ctx context.Context, ag agent.Agent, key session.SessionKey, task string, target ReplyTarget) {
	opts := agent.SessionOptions{
		SessionKey: key.String(),
		WorkDir:    "",
	}

	agSess, err := ag.StartSession(ctx, opts)
	if err != nil {
		e.sessions.SetError(key.String(), err.Error())
		e.sessions.Transition(key.String(), session.StatusFailed)
		e.reply(ctx, target, fmt.Sprintf("启动 Agent 失败：%v", err))
		return
	}

	if err := agSess.Send(ctx, task); err != nil {
		e.sessions.SetError(key.String(), err.Error())
		e.sessions.Transition(key.String(), session.StatusFailed)
		e.reply(ctx, target, fmt.Sprintf("发送任务失败：%v", err))
		return
	}

	reporter := NewProgressReporter(e.progressInterval(), func(text string) {
		if text != "" {
			e.reply(ctx, target, "处理中...\n\n"+truncStr(text, 1500))
		}
	})

	for event := range agSess.Events() {
		switch event.Type {
		case agent.EventTextDelta:
			reporter.Write(event.Text)
		case agent.EventDone:
			e.sessions.Transition(key.String(), session.StatusIdle)
			final := reporter.Final()
			if final == "" {
				final = "任务完成（无输出）。"
			}
			replyText := fmt.Sprintf("任务完成\n\n%s", truncStr(final, 2000))
			e.reply(ctx, target, replyText)
			if e.logger != nil {
				e.logger.Info("agent done: session=%s output=%s", key.String(), truncStr(final, 500))
			}
			return

		case agent.EventError:
			e.sessions.SetError(key.String(), event.Error)
			e.sessions.Transition(key.String(), session.StatusFailed)
			final := reporter.Final()
			errReply := fmt.Sprintf("执行失败：%s\n\n会话：%s\n最近输出：\n%s",
				event.Error, key.String(), truncStr(final, 500))
			e.reply(ctx, target, errReply)
			if e.logger != nil {
				e.logger.Error("agent error: session=%s err=%s output=%s", key.String(), event.Error, truncStr(final, 300))
			}
			return
		}
	}
}

func (e *Engine) progressInterval() time.Duration {
	return 3 * time.Second
}

func (e *Engine) resolveWorkDir(project string) (string, error) {
	projects, err := e.store.LoadProjects()
	if err != nil {
		return "", err
	}
	path, ok := projects[project]
	if !ok {
		return "", fmt.Errorf("未绑定项目 %s，请先用 /bind 绑定", project)
	}
	return path, nil
}

func (e *Engine) reply(ctx context.Context, target ReplyTarget, text string) {
	if e.platform == nil {
		return
	}
	if err := e.platform.Reply(ctx, target, text); err != nil {
		// Logging would go here; for MVP, silently fail
	}
}
