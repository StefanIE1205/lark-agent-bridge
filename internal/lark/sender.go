package lark

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const postThreshold = 500 // Use post format for messages longer than this

func (a *Adapter) sendMessage(ctx context.Context, target core.ReplyTarget, text string) (string, error) {
	// Use post format for long messages
	if len(text) > postThreshold {
		return a.sendPost(ctx, target, text)
	}
	return a.sendText(ctx, target, text)
}

func (a *Adapter) sendText(ctx context.Context, target core.ReplyTarget, text string) (string, error) {
	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return "", fmt.Errorf("lark: marshal content: %w", err)
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
		return "", fmt.Errorf("lark: create message: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("lark: create message failed: code=%d msg=%s",
			resp.Code, resp.Msg)
	}

	var msgID string
	if resp.Data != nil && resp.Data.MessageId != nil {
		msgID = *resp.Data.MessageId
		a.logger.Printf("reply sent: msg_id=%s chat=%s", msgID, target.ChatID)
	}

	return msgID, nil
}

func (a *Adapter) updateMessage(ctx context.Context, messageID string, text string) error {
	if a.apiClient == nil {
		return fmt.Errorf("lark: apiClient not initialized")
	}

	content, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("lark: marshal content: %w", err)
	}

	req := larkim.NewPatchMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewPatchMessageReqBodyBuilder().
			Content(string(content)).
			Build()).
		Build()

	resp, err := a.apiClient.Im.V1.Message.Patch(ctx, req)
	if err != nil {
		return fmt.Errorf("lark: update message: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("lark: update message failed: code=%d msg=%s",
			resp.Code, resp.Msg)
	}

	a.logger.Printf("message updated: msg_id=%s", messageID)
	return nil
}

func (a *Adapter) sendPost(ctx context.Context, target core.ReplyTarget, text string) (string, error) {
	// Split text into paragraphs
	lines := strings.Split(text, "\n")

	// Build post content - each line is a separate paragraph
	postContent := larkim.NewMessagePostContent().ContentTitle("")
	for _, line := range lines {
		if line == "" {
			continue
		}
		postContent.AppendContent([]larkim.MessagePostElement{
			&larkim.MessagePostText{Text: line},
		})
	}
	postContent.Build()

	// Build post message
	postMsg, err := larkim.NewMessagePost().
		ZhCn(postContent).
		Build()
	if err != nil {
		// Fallback to text
		return a.sendText(ctx, target, text)
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(target.ChatID).
			MsgType("post").
			Content(postMsg).
			Build()).
		Build()

	resp, err := a.apiClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		// Fallback to text
		return a.sendText(ctx, target, text)
	}

	if !resp.Success() {
		// Fallback to text
		return a.sendText(ctx, target, text)
	}

	var msgID string
	if resp.Data != nil && resp.Data.MessageId != nil {
		msgID = *resp.Data.MessageId
		a.logger.Printf("post reply sent: msg_id=%s chat=%s", msgID, target.ChatID)
	}

	return msgID, nil
}
