package db

import (
	"fmt"
	"log"

	supabase "github.com/supabase-community/supabase-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var SupabaseClient *supabase.Client

// InitDB initializes the database connection (GORM - deprecated, use InitSupabase)
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	//// Auto-migrate the schema
	//if err := DB.AutoMigrate(&Image{}); err != nil {
	//	return fmt.Errorf("failed to migrate database: %w", err)
	//}

	log.Println("Database schema migrated successfully")
	return nil
}

// InitSupabase initializes the Supabase client
func InitSupabase(supabaseURL, supabaseKey string) error {
	var err error
	SupabaseClient, err = supabase.NewClient(supabaseURL, supabaseKey, &supabase.ClientOptions{})
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	log.Println("Supabase client initialized successfully")
	return nil
}

// GetDB returns the database instance (deprecated)
func GetDB() *gorm.DB {
	return DB
}

// GetSupabaseClient returns the Supabase client instance
func GetSupabaseClient() *supabase.Client {
	return SupabaseClient
}
