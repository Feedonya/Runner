package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Feedonya/Runner/backend/filesctl"
	"github.com/Feedonya/Runner/backend/handler"
	"github.com/Feedonya/Runner/backend/model"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	ctx := context.Background()

	// Initialize MinIO client
	minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Initialize buckets
	buckets := []string{"code", "tests", "executables"}
	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			log.Fatalf("Failed to check bucket %s: %v", bucket, err)
		}
		if !exists {
			err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				log.Fatalf("Failed to create bucket %s: %v", bucket, err)
			}
			log.Printf("Created bucket %s", bucket)
		}
	}

	filesManager := filesctl.NewMinioManager(minioClient)

	// Initialize DragonFly (Redis-compatible) client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("DRAGONFLY_HOST"), os.Getenv("DRAGONFLY_PORT")),
		Password: os.Getenv("DRAGONFLY_PASSWORD"),
		DB:       0,
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to DragonFly: %v", err)
	}

	// Start pub/sub listener for task results
	go func() {
		pubsub := redisClient.Subscribe(ctx, "coderunner_completed_tasks_channel")
		for msg := range pubsub.Channel() {
			var task model.Task
			if err := json.Unmarshal([]byte(msg.Payload), &task); err != nil {
				log.Printf("Error unmarshaling task result: %v", err)
				continue
			}
			// Store task result in DragonFly
			jsonBytes, err := json.Marshal(task)
			if err != nil {
				log.Printf("Error marshaling task: %v", err)
				continue
			}
			redisClient.Set(ctx, fmt.Sprintf("task:%s", task.ID), string(jsonBytes), 0)
		}
	}()

	// Set up HTTP router
	router := mux.NewRouter()
	taskHandler := handler.NewTaskHandler(filesManager, redisClient)
	router.HandleFunc("/api/tasks", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/api/tasks/{id}", taskHandler.GetTask).Methods("GET")
	router.HandleFunc("/api/tasks", taskHandler.ListTasks).Methods("GET")

	// Add CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost", "http://127.0.0.1"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
