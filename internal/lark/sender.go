package lark

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (a *Adapter) sendMessage(ctx context.Context, target core.ReplyTarget, text string) error {
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("lark: marshal content: %w", err)
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(target.ChatID).
			MsgType("text").
			Content(string(content)).
			Build()).
		Build()

	resp, err := a.apiClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("lark: create message: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("lark: create message failed: code=%d msg=%s",
			resp.Code, resp.Msg)
	}

	if resp.Data != nil && resp.Data.MessageId != nil {
		a.logger.Printf("reply sent: msg_id=%s chat=%s", *resp.Data.MessageId, target.ChatID)
	}

	return nil
}
