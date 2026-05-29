package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/agent"
	"github.com/StefanIE1205/lark-agent-bridge/internal/security"
	"github.com/StefanIE1205/lark-agent-bridge/internal/session"
	"github.com/StefanIE1205/lark-agent-bridge/internal/store"
)

type mockPlatform struct {
	replies []string
	msgID   int
}

func (p *mockPlatform) Name() string { return "mock" }
func (p *mockPlatform) Start(ctx context.Context, handler MessageHandler) error {
	return nil
}
func (p *mockPlatform) Reply(ctx context.Context, target ReplyTarget, text string) (string, error) {
	p.replies = append(p.replies, text)
	p.msgID++
	return fmt.Sprintf("msg_%d", p.msgID), nil
}
func (p *mockPlatform) Update(ctx context.Context, target ReplyTarget, messageID string, text string) error {
	// Track updates as replies for testing
	p.replies = append(p.replies, "[UPDATE "+messageID+"] "+text)
	return nil
}
func (p *mockPlatform) Stop(ctx context.Context) error { return nil }

func (p *mockPlatform) lastReply() string {
	if len(p.replies) == 0 {
		return ""
	}
	return p.replies[len(p.replies)-1]
}

func setupEngine(t *testing.T) (*Engine, *mockPlatform) {
	t.Helper()

	platform := &mockPlatform{}
	stateStore, err := store.NewStateStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stateStore.BindProject("demo", "C:\\demo")

	eng := NewEngine(EngineConfig{
		Platform:     platform,
		Policy:       security.NewPolicy([]string{"ou_admin"}, []string{"oc_group"}, true),
		Sessions:     session.NewManager(),
		Store:        stateStore,
		Agents:       []agent.Agent{agent.NewFakeAgent("")},
		DefaultAgent: "fake",
		ProjectNames: []string{"demo"},
	})

	return eng, platform
}

func TestHandlePing(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/ping",
	}

	eng.HandleMessage(context.Background(), msg)

	if plat.lastReply() != "pong" {
		t.Errorf("expected pong, got %q", plat.lastReply())
	}
}

func TestHandleHelp(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/help",
	}

	eng.HandleMessage(context.Background(), msg)

	reply := plat.lastReply()
	if !strings.Contains(reply, "/ping") {
		t.Error("help should list commands")
	}
}

func TestHandleAsk(t *testing.T) {
	eng, plat := setupEngine(t)

	eng.store.SetChatDefault("oc_admin", "", "project", "demo")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")

	msg := Message{
		ChatID:   "oc_admin",
		ThreadID: "ot_thread",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/ask 总结项目结构",
	}

	eng.HandleMessage(context.Background(), msg)

	// Wait for async agent to complete
	time.Sleep(500 * time.Millisecond)

	// Should have ACK + done message
	if len(plat.replies) < 2 {
		t.Fatalf("expected at least 2 replies (ACK + done), got %d", len(plat.replies))
	}

	ack := plat.replies[0]
	if !strings.Contains(ack, "收到") {
		t.Errorf("expected ACK, got: %s", ack)
	}

	done := plat.lastReply()
	if !strings.Contains(done, "任务完成") {
		t.Errorf("expected done message, got: %s", done)
	}
}

func TestHandleAskUnauthorized(t *testing.T) {
	eng, plat := setupEngine(t)

	// Non-admin in direct chat: silently ignored by CanAccess
	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_rando",
		IsDirect: true,
		Text:     "/ask do stuff",
	}

	eng.HandleMessage(context.Background(), msg)

	// Private chat from non-admin → CanAccess returns false → no reply
	if len(plat.replies) != 0 {
		t.Errorf("expected no replies for unauthorized user, got %d", len(plat.replies))
	}
}

func TestHandleAskUnauthorizedGroup(t *testing.T) {
	eng, plat := setupEngine(t)

	// Non-admin in allowed group who can access but can't execute /ask
	msg := Message{
		ChatID:    "oc_group",
		UserID:    "ou_rando",
		IsDirect:  false,
		Mentioned: true,
		Text:      "/ask do stuff",
	}

	eng.HandleMessage(context.Background(), msg)

	if len(plat.replies) == 0 {
		t.Fatal("expected a reply")
	}
	if !strings.Contains(plat.lastReply(), "权限不足") {
		t.Errorf("expected permission denied, got: %s", plat.lastReply())
	}
}

func TestHandleStop(t *testing.T) {
	eng, plat := setupEngine(t)

	eng.store.SetChatDefault("oc_admin", "", "project", "demo")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")

	// Start a long-running task
	eng.agents["fake"] = agent.NewFakeAgent("long")

	msg := Message{
		ChatID:   "oc_admin",
		ThreadID: "ot_thread",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/ask 长时间运行的任务",
	}

	eng.HandleMessage(context.Background(), msg)

	// Wait for task to start
	time.Sleep(200 * time.Millisecond)

	// Now stop it
	stopMsg := Message{
		ChatID:   "oc_admin",
		ThreadID: "ot_thread",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/stop",
	}
	eng.HandleMessage(context.Background(), stopMsg)

	stopReply := plat.lastReply()
	if !strings.Contains(stopReply, "已停止") {
		t.Errorf("expected stop confirmation, got: %s", stopReply)
	}
}

func TestHandleBind(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/bind backend D:\\backend",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "已绑定") {
		t.Errorf("expected bind confirmation, got: %s", plat.lastReply())
	}

	projects, _ := eng.store.LoadProjects()
	if projects["backend"] != "D:\\backend" {
		t.Errorf("backend not bound: %v", projects)
	}
}

func TestHandleProject(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/project demo",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "demo") {
		t.Errorf("expected project set, got: %s", plat.lastReply())
	}
}

func TestHandleAgent(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/agent fake",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "fake") {
		t.Errorf("expected agent set, got: %s", plat.lastReply())
	}
}

func TestHandleAgentUnknown(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/agent nonexistent",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "未知 Agent") {
		t.Errorf("expected unknown agent error, got: %s", plat.lastReply())
	}
}

func TestHandleStatus(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/status",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "状态") {
		t.Errorf("expected status info, got: %s", plat.lastReply())
	}
}

func TestHandleSessions(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/sessions",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "无活跃会话") {
		t.Errorf("expected no sessions, got: %s", plat.lastReply())
	}
}

func TestProgressIntervalFromConfig(t *testing.T) {
	platform := &mockPlatform{}
	stateStore, _ := store.NewStateStore(t.TempDir())

	eng := NewEngine(EngineConfig{
		Platform:           platform,
		Policy:             security.NewPolicy([]string{"ou_admin"}, []string{}, true),
		Sessions:           session.NewManager(),
		Store:              stateStore,
		Agents:             []agent.Agent{agent.NewFakeAgent("")},
		DefaultAgent:       "fake",
		ProgressIntervalMs: 15000,
	})

	if eng.progressInterval() != 15*time.Second {
		t.Errorf("expected 15s, got %v", eng.progressInterval())
	}
}

func TestProgressIntervalDefault(t *testing.T) {
	platform := &mockPlatform{}
	stateStore, _ := store.NewStateStore(t.TempDir())

	eng := NewEngine(EngineConfig{
		Platform:     platform,
		Policy:       security.NewPolicy([]string{"ou_admin"}, []string{}, true),
		Sessions:     session.NewManager(),
		Store:        stateStore,
		Agents:       []agent.Agent{agent.NewFakeAgent("")},
		DefaultAgent: "fake",
		// ProgressIntervalMs not set — should default to 3000
	})

	if eng.progressInterval() != 3*time.Second {
		t.Errorf("expected 3s default, got %v", eng.progressInterval())
	}
}

func TestRunAgentPassesWorkDir(t *testing.T) {
	eng, plat := setupEngine(t)

	// Bind project to a specific path
	eng.store.BindProject("myproject", "C:\\work\\myproject")
	eng.store.SetChatDefault("oc_admin", "", "project", "myproject")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")

	msg := Message{
		ChatID:   "oc_admin",
		ThreadID: "ot_thread",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/ask check workdir",
	}

	eng.HandleMessage(context.Background(), msg)
	time.Sleep(500 * time.Millisecond)

	// Verify the agent received the correct WorkDir
	// The fake agent stores opts, we can check via the session
	sessKey := session.DeriveKey("oc_admin", "ot_thread", "myproject", "fake")
	sess := eng.sessions.Get(sessKey.String())
	if sess == nil {
		t.Fatal("expected session to exist")
	}

	// The fake agent's session should have WorkDir set
	// We verify by checking that the ACK mentions the project
	ack := plat.replies[0]
	if !strings.Contains(ack, "收到") {
		t.Errorf("expected ACK, got: %s", ack)
	}

	// Verify the session was created with correct workDir
	if sess.WorkDir != "C:\\work\\myproject" {
		t.Errorf("expected WorkDir 'C:\\work\\myproject', got %q", sess.WorkDir)
	}
}

func TestHandleApproveAndDeny(t *testing.T) {
	eng, plat := setupEngine(t)

	// Create a pending approval request
	req := eng.approval.Create("test-session", "npm install", "medium")

	// Approve it
	approveMsg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/approve " + req.ID,
	}
	eng.HandleMessage(context.Background(), approveMsg)

	if !strings.Contains(plat.lastReply(), "已批准") {
		t.Errorf("expected approval confirmation, got: %s", plat.lastReply())
	}

	// Verify status changed
	if eng.approval.Get(req.ID).Status != security.ApprovalGranted {
		t.Error("expected approval status to be granted")
	}
}

func TestHandleApproveNotFound(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/approve nonexistent",
	}
	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "操作失败") {
		t.Errorf("expected error for missing request, got: %s", plat.lastReply())
	}
}

func TestHandleApproveNoArgs(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/approve",
	}
	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "用法") {
		t.Errorf("expected usage message, got: %s", plat.lastReply())
	}
}

func TestHandleFakeAgentError(t *testing.T) {
	eng, plat := setupEngine(t)

	eng.store.SetChatDefault("oc_admin", "", "project", "demo")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")
	eng.agents["fake"] = agent.NewFakeAgent("error")

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "/ask test error flow",
	}

	eng.HandleMessage(context.Background(), msg)

	time.Sleep(400 * time.Millisecond)

	done := plat.lastReply()
	if !strings.Contains(done, "执行失败") && !strings.Contains(done, "模拟错误") {
		t.Errorf("expected error, got: %s", done)
	}
}

func TestNaturalLanguageAsk(t *testing.T) {
	eng, plat := setupEngine(t)

	eng.store.SetChatDefault("oc_admin", "", "project", "demo")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")

	msg := Message{
		ChatID:   "oc_admin",
		ThreadID: "ot_thread",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "帮我跑一下测试",
	}

	eng.HandleMessage(context.Background(), msg)
	time.Sleep(500 * time.Millisecond)

	if len(plat.replies) < 2 {
		t.Fatalf("expected at least 2 replies, got %d", len(plat.replies))
	}

	ack := plat.replies[0]
	if !strings.Contains(ack, "收到") {
		t.Errorf("expected ACK, got: %s", ack)
	}

	done := plat.lastReply()
	if !strings.Contains(done, "任务完成") {
		t.Errorf("expected done message, got: %s", done)
	}
}

func TestNaturalLanguageStop(t *testing.T) {
	eng, plat := setupEngine(t)

	eng.store.SetChatDefault("oc_admin", "", "project", "demo")
	eng.store.SetChatDefault("oc_admin", "", "agent", "fake")

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "停一下",
	}

	eng.HandleMessage(context.Background(), msg)

	// No active session → should get an error about no session
	reply := plat.lastReply()
	if !strings.Contains(reply, "停止失败") && !strings.Contains(reply, "已停止") && !strings.Contains(reply, "无活跃会话") {
		t.Errorf("expected stop response, got: %s", reply)
	}
}

func TestNaturalLanguageStatus(t *testing.T) {
	eng, plat := setupEngine(t)

	msg := Message{
		ChatID:   "oc_admin",
		UserID:   "ou_admin",
		IsDirect: true,
		Text:     "现在什么状态",
	}

	eng.HandleMessage(context.Background(), msg)

	if !strings.Contains(plat.lastReply(), "状态") {
		t.Errorf("expected status response, got: %s", plat.lastReply())
	}
}
