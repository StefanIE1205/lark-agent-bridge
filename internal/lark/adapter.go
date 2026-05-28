package lark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	platformName = "lark"
	dedupTTL     = 5 * time.Minute
	dedupCleanup = 1 * time.Minute
)

type dedupEntry struct {
	seen time.Time
}

type Adapter struct {
	appID     string
	appSecret string
	domain    string
	botOpenID string
	botName   string
	logger    *log.Logger
	wsClient  *larkws.Client
	apiClient *lark.Client
	dedup     map[string]dedupEntry
	dedupMu   sync.Mutex
}

func NewAdapter(appID, appSecret, domain string, logger *log.Logger) *Adapter {
	if logger == nil {
		logger = log.New(os.Stderr, "[lark] ", log.LstdFlags)
	}
	if domain == "" {
		domain = "https://open.feishu.cn"
	}
	return &Adapter{
		appID:     appID,
		appSecret: appSecret,
		domain:    domain,
		logger:    logger,
		dedup:     make(map[string]dedupEntry),
	}
}

func (a *Adapter) Name() string {
	return platformName
}

func (a *Adapter) Start(ctx context.Context, handler core.MessageHandler) error {
	a.apiClient = lark.NewClient(a.appID, a.appSecret,
		lark.WithOpenBaseUrl(a.domain),
	)

	// Discover bot identity (like cc-connect)
	a.discoverBotInfo(ctx)

	dispatch := dispatcher.NewEventDispatcher("", "")
	dispatch.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		a.handleMessage(ctx, event, handler)
		return nil
	})

	a.wsClient = larkws.NewClient(a.appID, a.appSecret,
		larkws.WithEventHandler(dispatch),
		larkws.WithAutoReconnect(true),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
		larkws.WithDomain(a.domain),
	)

	go a.dedupGC(ctx)

	a.logger.Printf("lark adapter starting (app_id=%s..., domain=%s, bot=%s)", a.appID[:min(8, len(a.appID))], a.domain, a.botOpenID)

	if err := a.wsClient.Start(ctx); err != nil {
		return fmt.Errorf("lark: start client: %w", err)
	}

	return nil
}

func (a *Adapter) discoverBotInfo(ctx context.Context) {
	// Get tenant access token
	tokenResp, err := a.apiClient.GetTenantAccessTokenBySelfBuiltApp(ctx,
		&larkcore.SelfBuiltTenantAccessTokenReq{AppID: a.appID, AppSecret: a.appSecret})
	if err != nil {
		a.logger.Printf("warn: get tenant token for bot info: %v", err)
		return
	}
	if tokenResp == nil || tokenResp.TenantAccessToken == "" {
		a.logger.Printf("warn: empty tenant token response")
		return
	}

	// Call bot info API
	req, _ := http.NewRequestWithContext(ctx, "GET",
		a.domain+"/open-apis/bot/v3/info", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.TenantAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.logger.Printf("warn: bot info request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var botResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Bot  struct {
			OpenID      string `json:"open_id"`
			AppName     string `json:"app_name"`
			ActivateStatus int `json:"activate_status"`
		} `json:"bot"`
	}
	if err := json.Unmarshal(body, &botResp); err != nil {
		a.logger.Printf("warn: parse bot info: %v", err)
		return
	}

	if botResp.Bot.OpenID != "" {
		a.botOpenID = botResp.Bot.OpenID
	}
	if botResp.Bot.AppName != "" {
		a.botName = botResp.Bot.AppName
	}
	a.logger.Printf("bot discovered: open_id=%s name=%s", a.botOpenID, a.botName)
}

func (a *Adapter) Stop(ctx context.Context) error {
	if a.wsClient != nil {
		a.wsClient.Close()
		a.logger.Printf("lark adapter stopped")
	}
	return nil
}

func (a *Adapter) Reply(ctx context.Context, target core.ReplyTarget, text string) error {
	return a.sendMessage(ctx, target, text)
}

func (a *Adapter) Update(ctx context.Context, target core.ReplyTarget, messageID string, text string) error {
	return fmt.Errorf("lark: Update not implemented")
}

func (a *Adapter) handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1, handler core.MessageHandler) {
	msg := a.parseMessage(event)

	// Debug: log raw event
	if event.Event != nil && event.Event.Message != nil {
		m := event.Event.Message
		a.logger.Printf("RAW EVENT: msg_id=%v chat_id=%v chat_type=%v msg_type=%v",
			strPtr(m.MessageId), strPtr(m.ChatId), strPtr(m.ChatType), strPtr(m.MessageType))
	}

	// Filter out bot's own messages (like cc-connect)
	if a.botOpenID != "" && msg.UserID == a.botOpenID {
		a.logger.Printf("skipping self-message from bot (open_id=%s)", a.botOpenID)
		return
	}

	if msg.ID != "" && a.isDuplicate(msg.ID) {
		a.logger.Printf("skipping duplicate message: id=%s", msg.ID)
		return
	}

	a.logger.Printf("RECEIVED: id=%s chat=%s user=%s text=%q platform=%s direct=%v",
		msg.ID, msg.ChatID, msg.UserID, truncate(msg.Text, 80), msg.Platform, msg.IsDirect)

	if handler != nil {
		handler(ctx, msg)
	}
}

func (a *Adapter) isDuplicate(id string) bool {
	a.dedupMu.Lock()
	defer a.dedupMu.Unlock()

	if _, exists := a.dedup[id]; exists {
		return true
	}
	a.dedup[id] = dedupEntry{seen: time.Now()}
	return false
}

func (a *Adapter) dedupGC(ctx context.Context) {
	ticker := time.NewTicker(dedupCleanup)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.purgeOldEntries()
		}
	}
}

func (a *Adapter) purgeOldEntries() {
	a.dedupMu.Lock()
	defer a.dedupMu.Unlock()

	cutoff := time.Now().Add(-dedupTTL)
	for id, entry := range a.dedup {
		if entry.seen.Before(cutoff) {
			delete(a.dedup, id)
		}
	}
}
