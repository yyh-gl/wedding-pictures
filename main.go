package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yyh-gl/wedding-picture/internal/db"
	"github.com/yyh-gl/wedding-picture/internal/handler"
	"github.com/yyh-gl/wedding-picture/internal/storage"
)

func main() {
	// Load environment variables
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	awsS3Bucket := os.Getenv("AWS_S3_BUCKET")
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Validate required environment variables
	if channelSecret == "" || channelToken == "" {
		log.Fatalf("environment variables LINE_CHANNEL_SECRET and LINE_CHANNEL_ACCESS_TOKEN must be set")
	}

	if supabaseURL == "" || supabaseKey == "" {
		log.Fatalf("environment variables SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}

	if awsRegion == "" || awsS3Bucket == "" || awsAccessKeyID == "" || awsSecretAccessKey == "" {
		log.Fatalf("environment variables AWS_REGION, AWS_S3_BUCKET, AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set")
	}

	// Initialize Supabase client
	if err := db.InitSupabase(supabaseURL, supabaseKey); err != nil {
		log.Fatalf("failed to initialize Supabase: %v", err)
	}

	log.Println("Supabase client initialized successfully")

	// Initialize S3 service
	ctx := context.Background()

	s3Service, err := storage.NewS3Service(ctx, awsRegion, awsS3Bucket, awsAccessKeyID, awsSecretAccessKey)
	if err != nil {
		log.Fatalf("failed to initialize S3 service: %v", err)
	}

	log.Println("S3 service initialized successfully")

	// Initialize handlers
	webhookHandler, err := handler.NewWebhookHandler(channelSecret, channelToken, s3Service)
	if err != nil {
		log.Fatalf("failed to initialize webhook handler: %v", err)
	}

	imagesHandler := handler.NewImagesHandler()

	// Register routes
	http.HandleFunc("/callback", webhookHandler.Handle)
	http.HandleFunc("/api/images", imagesHandler.Handle)
	http.HandleFunc("/api/images/displayed", imagesHandler.MarkDisplayed)

	// Serve static files
	http.Handle("/", http.FileServer(http.Dir("./static")))

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
