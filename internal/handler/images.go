package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/supabase-community/postgrest-go"
	"github.com/yyh-gl/wedding-picture/internal/db"
)

type ImageResponse struct {
	ID        uint   `json:"id"`
	FileURL   string `json:"file_url"`
	IsNew     bool   `json:"is_new"`
	CreatedAt string `json:"created_at"`
}

type ImagesHandler struct{}

func NewImagesHandler() *ImagesHandler {
	return &ImagesHandler{}
}

// Handle returns images with the following priority:
// 1. New images first (is_new = true)
// 2. Then undisplayed images in upload order
// 3. If all images are displayed, reset all and return from the beginning
func (h *ImagesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	client := db.GetSupabaseClient()

	// Check if there are any new images
	var newImages []db.Image
	data, _, err := client.From("images").
		Select("*", "", false).
		Eq("is_new", "true").
		Order("uploaded_at", &postgrest.OrderOpts{Ascending: true}).
		Execute()
	if err != nil {
		log.Printf("failed to query new images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	if err := json.Unmarshal(data, &newImages); err != nil {
		log.Printf("failed to unmarshal new images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	// If there are new images, return them
	if len(newImages) > 0 {
		response := make([]ImageResponse, len(newImages))
		for i, img := range newImages {
			response[i] = ImageResponse{
				ID:        img.ID,
				FileURL:   img.FileURL,
				IsNew:     img.IsNew,
				CreatedAt: img.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check for undisplayed images
	var undisplayedImages []db.Image
	data, _, err = client.From("images").
		Select("*", "", false).
		Eq("is_displayed", "false").
		Order("uploaded_at", &postgrest.OrderOpts{Ascending: true}).
		Execute()
	if err != nil {
		log.Printf("failed to query undisplayed images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	if err := json.Unmarshal(data, &undisplayedImages); err != nil {
		log.Printf("failed to unmarshal undisplayed images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	// If there are undisplayed images, return them
	if len(undisplayedImages) > 0 {
		response := make([]ImageResponse, len(undisplayedImages))
		for i, img := range undisplayedImages {
			response[i] = ImageResponse{
				ID:        img.ID,
				FileURL:   img.FileURL,
				IsNew:     img.IsNew,
				CreatedAt: img.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// All images have been displayed, reset and return all images
	updateData := map[string]interface{}{
		"is_displayed": false,
	}
	_, _, err = client.From("images").
		Update(updateData, "", "").
		Eq("is_displayed", "true").
		Execute()
	if err != nil {
		log.Printf("failed to reset displayed status: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	log.Println("All images displayed, resetting display status")

	// Return all images in upload order
	var allImages []db.Image
	data, _, err = client.From("images").
		Select("*", "", false).
		Order("uploaded_at", &postgrest.OrderOpts{Ascending: true}).
		Execute()
	if err != nil {
		log.Printf("failed to query all images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	if err := json.Unmarshal(data, &allImages); err != nil {
		log.Printf("failed to unmarshal all images: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	response := make([]ImageResponse, len(allImages))
	for i, img := range allImages {
		response[i] = ImageResponse{
			ID:        img.ID,
			FileURL:   img.FileURL,
			IsNew:     img.IsNew,
			CreatedAt: img.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MarkDisplayed marks an image as displayed
func (h *ImagesHandler) MarkDisplayed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		ID uint `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	client := db.GetSupabaseClient()

	// Mark as displayed and no longer new
	updateData := map[string]interface{}{
		"is_displayed": true,
		"is_new":       false,
	}
	_, _, err := client.From("images").
		Update(updateData, "", "").
		Eq("id", fmt.Sprintf("%d", req.ID)).
		Execute()
	if err != nil {
		log.Printf("failed to mark image as displayed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
