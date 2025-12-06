package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"github.com/yyh-gl/wedding-picture/internal/db"
	"github.com/yyh-gl/wedding-picture/internal/storage"
)

type WebhookHandler struct {
	channelSecret string
	channelToken  string
	client        *messaging_api.MessagingApiAPI
	s3Service     *storage.S3Service
}

func NewWebhookHandler(channelSecret, channelToken string, s3Service *storage.S3Service) (*WebhookHandler, error) {
	client, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LINE Messaging API client: %w", err)
	}

	return &WebhookHandler{
		channelSecret: channelSecret,
		channelToken:  channelToken,
		client:        client,
		s3Service:     s3Service,
	}, nil
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "method not allowed")
		return
	}

	reqBody, err := webhook.ParseRequest(h.channelSecret, r)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			log.Printf("signature verification failed: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad request")
			return
		}
		log.Printf("failed to parse webhook request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal error")
		return
	}

	if reqBodyJSON, err := json.MarshalIndent(reqBody, "", "  "); err == nil {
		log.Printf("Received webhook request:\n%s", string(reqBodyJSON))
	}

	for _, ev := range reqBody.Events {
		log.Printf("Processing event type: %T", ev)
		switch e := ev.(type) {
		case webhook.MessageEvent:
			h.handleMessageEvent(e)
		default:
			log.Printf("Unhandled event type: %T", ev)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (h *WebhookHandler) handleMessageEvent(e webhook.MessageEvent) {
	switch msg := e.Message.(type) {
	case webhook.TextMessageContent:
		log.Printf("TextMessage - ID: %s, Text: %s", msg.Id, msg.Text)
		h.replyText(e.ReplyToken, msg.Text)

	case webhook.ImageMessageContent:
		log.Printf("ImageMessage - ID: %s", msg.Id)
		h.handleImageMessage(e.ReplyToken, msg.Id, e.Source)

	default:
		log.Printf("Unhandled message type: %T", msg)
	}
}

func (h *WebhookHandler) handleImageMessage(replyToken, messageID string, source webhook.SourceInterface) {
	ctx := context.Background()

	// Get uploader name from LINE profile
	uploaderName := h.getUploaderName(source)

	// Get image content from LINE using blob API
	blobAPI, err := messaging_api.NewMessagingApiBlobAPI(h.channelToken)
	if err != nil {
		log.Printf("failed to initialize blob API: %v", err)
		h.replyText(replyToken, "画像の取得に失敗しました")
		return
	}

	contentResp, err := blobAPI.GetMessageContent(messageID)
	if err != nil {
		log.Printf("failed to get image content: %v", err)
		h.replyText(replyToken, "画像の取得に失敗しました")
		return
	}
	defer contentResp.Body.Close()

	// Read image data
	imageData, err := io.ReadAll(contentResp.Body)
	if err != nil {
		log.Printf("failed to read image data: %v", err)
		h.replyText(replyToken, "画像の読み込みに失敗しました")
		return
	}

	// Upload to S3
	s3Key := fmt.Sprintf("image_%s_%d.jpg", messageID, time.Now().Unix())
	imageFile, err := h.s3Service.UploadImage(ctx, s3Key, "image/jpeg", bytes.NewReader(imageData))
	if err != nil {
		log.Printf("failed to upload to S3: %v", err)
		h.replyText(replyToken, "S3へのアップロードに失敗しました")
		return
	}

	// Save to database
	image := map[string]interface{}{
		"s3_key":        imageFile.Key,
		"file_url":      imageFile.URL,
		"uploaded_at":   imageFile.UploadedAt,
		"is_new":        true,
		"is_displayed":  false,
		"uploader_name": uploaderName,
	}

	_, _, err = db.GetSupabaseClient().
		From("images").
		Insert(image, false, "", "", "").
		Execute()
	if err != nil {
		log.Printf("failed to save image to database: %v", err)
		h.replyText(replyToken, "データベースへの保存に失敗しました")
		return
	}

	log.Printf("Image uploaded successfully: %s (S3 Key: %s, Uploader: %s)", imageFile.Name, imageFile.Key, uploaderName)
	h.replyText(replyToken, "写真をアップロードしました!")
}

func (h *WebhookHandler) getUploaderName(source webhook.SourceInterface) string {
	switch s := source.(type) {
	case webhook.UserSource:
		profile, err := h.client.GetProfile(s.UserId)
		if err != nil {
			log.Printf("failed to get user profile: %v", err)
			return "Unknown"
		}
		return profile.DisplayName
	case webhook.GroupSource:
		profile, err := h.client.GetGroupMemberProfile(s.GroupId, s.UserId)
		if err != nil {
			log.Printf("failed to get group member profile: %v", err)
			return "Unknown"
		}
		return profile.DisplayName
	case webhook.RoomSource:
		profile, err := h.client.GetRoomMemberProfile(s.RoomId, s.UserId)
		if err != nil {
			log.Printf("failed to get room member profile: %v", err)
			return "Unknown"
		}
		return profile.DisplayName
	default:
		log.Printf("unknown source type: %T", source)
		return "Unknown"
	}
}

func (h *WebhookHandler) replyText(replyToken, text string) {
	_, err := h.client.ReplyMessage(&messaging_api.ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []messaging_api.MessageInterface{
			&messaging_api.TextMessage{
				Text: text,
			},
		},
	})
	if err != nil {
		log.Printf("failed to reply: %v", err)
	}
}
