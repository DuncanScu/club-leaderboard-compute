package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Duncanscu/leaderboard-computer/services"
)

// Global variables for connection reuse across Lambda invocations
var (
	db                   *gorm.DB
	leaderboardService   services.LeaderboardComputeService
)

// handler processes EventBridge scheduled events to compute leaderboards
func handler(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("[LAMBDA] Computing leaderboards at %s", event.Time)

	// Initialize database connection on cold start
	if db == nil {
		if err := initDB(ctx); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	// Compute all leaderboard types
	if err := leaderboardService.ComputeAllLeaderboards(ctx); err != nil {
		log.Printf("[LAMBDA_ERROR] Failed to compute leaderboards: %v", err)
		return err
	}

	log.Printf("[LAMBDA_SUCCESS] Successfully computed leaderboards")
	return nil
}

// initDB initializes the database connection
func initDB(ctx context.Context) error {
	log.Println("[LAMBDA] Initializing database connection...")

	// Get database credentials from environment or Secrets Manager
	var dbHost, dbPort, dbUser, dbPassword, dbName string

	dbSecretArn := os.Getenv("DB_SECRET_ARN")
	if dbSecretArn != "" {
		// Fetch from Secrets Manager
		creds, err := getDBCredentials(ctx, dbSecretArn)
		if err != nil {
			return fmt.Errorf("failed to get DB credentials: %w", err)
		}
		dbHost = creds["host"]
		dbPort = creds["port"]
		dbUser = creds["username"]
		dbPassword = creds["password"]
		dbName = creds["dbname"]
	} else {
		// Fallback to environment variables
		dbHost = os.Getenv("DB_HOST")
		dbPort = os.Getenv("DB_PORT")
		dbUser = os.Getenv("DB_USER")
		dbPassword = os.Getenv("DB_PASSWORD")
		dbName = os.Getenv("DB_NAME")
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for Lambda
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Lambda-optimized connection pooling
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)

	// Initialize services
	leaderboardService = services.NewLeaderboardComputeService(db)

	log.Println("[LAMBDA] Database initialized successfully")
	return nil
}

// getDBCredentials fetches database credentials from Secrets Manager
func getDBCredentials(ctx context.Context, secretArn string) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretArn,
	})
	if err != nil {
		return nil, err
	}

	var creds map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func main() {
	lambda.Start(handler)
}
