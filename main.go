package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func main() {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")

	if channelSecret == "" || channelToken == "" {
		log.Fatalf("environment variables LINE_CHANNEL_SECRET and LINE_CHANNEL_ACCESS_TOKEN must be set")
	}

	client, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		log.Fatalf("failed to initialize LINE Messaging API client: %v", err)
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "method not allowed")
			return
		}

		// Verify signature and parse events using webhook package
		reqBody, err := webhook.ParseRequest(channelSecret, r)
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

		// Log the entire request body
		if reqBodyJSON, err := json.MarshalIndent(reqBody, "", "  "); err == nil {
			log.Printf("Received webhook request:\n%s", string(reqBodyJSON))
		} else {
			log.Printf("Received webhook request (JSON marshal failed): %+v", reqBody)
		}

		for _, ev := range reqBody.Events {
			log.Printf("Processing event type: %T", ev)
			switch e := ev.(type) {
			case *webhook.MessageEvent:
				log.Printf("MessageEvent - Type: %s, ReplyToken: %s", e.Type, e.ReplyToken)
				// Only handle text message events
				switch msg := e.Message.(type) {
				case *webhook.TextMessageContent:
					log.Printf("TextMessage - ID: %s, Text: %s", msg.Id, msg.Text)
					_, err := client.ReplyMessage(&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							&messaging_api.TextMessage{
								Text: "hello",
							},
						},
					})
					if err != nil {
						log.Printf("failed to reply: %v", err)
					}
				default:
					// ignore non-text messages
				}
			default:
				// ignore non-message events
			}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "method not allowed")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
