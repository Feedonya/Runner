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

	// Start pub/sub listener for task results with enhanced debugging
	go func() {
		pubsub := redisClient.Subscribe(ctx, "coderunner_completed_tasks_channel")
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			log.Printf("Received raw message from channel: %s", msg.Payload)

			// First attempt: Unmarshal into model.Task
			var task model.Task
			if err := json.Unmarshal([]byte(msg.Payload), &task); err != nil {
				log.Printf("Error unmarshaling into model.Task: %v", err)
			} else {
				log.Printf("Successfully unmarshaled task: %+v", task)
			}

			// Fallback: Parse as generic JSON to handle unexpected format
			var rawData map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &rawData); err != nil {
				log.Printf("Error unmarshaling into generic map: %v", err)
				continue
			}
			log.Printf("Raw data structure: %+v", rawData)

			// Extract and map fields
			if id, ok := rawData["id"].(string); ok {
				task.ID = id
			}
			if state, ok := rawData["state"].(string); ok {
				task.State = state
			}
			// Override state to "completed" if testsResults is present
			if rawData["testsResults"] != nil {
				task.State = "completed"
				log.Printf("Overriding state to 'completed' due to presence of testsResults")
			}

			// Handle testsResults
			if results, ok := rawData["testsResults"].([]interface{}); ok {
				task.TestsResults = make([]model.TestResult, len(results))
				for i, r := range results {
					resultMap, ok := r.(map[string]interface{})
					if !ok {
						log.Printf("Invalid test result format at index %d", i)
						continue
					}
					// Map test_id to TestNumber and successful to Passed
					if testID, ok := resultMap["test_id"].(float64); ok {
						task.TestsResults[i] = model.TestResult{
							TestNumber: int(testID),
							Passed:     resultMap["successful"].(bool),
						}
					} else {
						// Fallback if test_id is missing
						if successful, ok := resultMap["successful"].(bool); ok {
							task.TestsResults[i] = model.TestResult{
								Passed: successful,
							}
						}
					}
				}
			}

			// Fetch existing task data to preserve fields
			existingTaskJSON, err := redisClient.Get(ctx, fmt.Sprintf("task:%s", task.ID)).Result()
			if err == nil {
				var existingTask model.Task
				if err := json.Unmarshal([]byte(existingTaskJSON), &existingTask); err == nil {
					if task.CodeLocation.BucketName == "" {
						task.CodeLocation = existingTask.CodeLocation
					}
					if task.TestsLocation.BucketName == "" {
						task.TestsLocation = existingTask.TestsLocation
					}
					if task.ExecutableLocation.BucketName == "" {
						task.ExecutableLocation = existingTask.ExecutableLocation
					}
					if task.Compiler == "" {
						task.Compiler = existingTask.Compiler
					}
				}
			}

			// Store task result in DragonFly
			jsonBytes, err := json.Marshal(task)
			if err != nil {
				log.Printf("Error marshaling task: %v", err)
				continue
			}
			err = redisClient.Set(ctx, fmt.Sprintf("task:%s", task.ID), string(jsonBytes), 0).Err()
			if err != nil {
				log.Printf("Error setting task in DragonFly: %v", err)
				continue
			}
			log.Printf("Updated task %s in DragonFly: %s", task.ID, string(jsonBytes))
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
