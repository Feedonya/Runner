package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Feedonya/Runner/backend/filesctl"
	"github.com/Feedonya/Runner/backend/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type TaskHandler struct {
	filesManager filesctl.Manager
	redisClient  *redis.Client
}

func NewTaskHandler(filesManager filesctl.Manager, redisClient *redis.Client) *TaskHandler {
	return &TaskHandler{
		filesManager: filesManager,
		redisClient:  redisClient,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get code file
	file, _, err := r.FormFile("code")
	if err != nil {
		http.Error(w, "Missing code file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	codeData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read code file", http.StatusBadRequest)
		return
	}

	// Get task ID and compiler
	taskID := r.FormValue("task_id")
	if taskID == "" {
		http.Error(w, "Missing task_id", http.StatusBadRequest)
		return
	}
	compiler := r.FormValue("compiler")
	if compiler == "" {
		compiler = "g++"
	}

	// Generate unique file name
	attemptID := uuid.New().String()
	codeObjectName := fmt.Sprintf("%s.cpp", attemptID)

	// Upload code to MinIO
	err = h.filesManager.PutFile(ctx, "code", codeObjectName, codeData)
	if err != nil {
		http.Error(w, "Failed to upload code to MinIO", http.StatusInternalServerError)
		return
	}

	// Create task command
	taskCommand := model.StartTaskCommand{
		ID: fmt.Sprintf("task_%s_%s", taskID, attemptID),
		CodeLocation: model.FileLocation{
			BucketName: "code",
			ObjectName: codeObjectName,
		},
		TestsLocation: model.FileLocation{
			BucketName: "tests",
			ObjectName: fmt.Sprintf("%s.json", taskID),
		},
		Compiler: compiler,
	}

	// Publish task to DragonFly
	jsonBytes, err := json.Marshal(taskCommand)
	if err != nil {
		http.Error(w, "Failed to marshal task", http.StatusInternalServerError)
		return
	}
	err = h.redisClient.Publish(ctx, "coderunner_task_channel", string(jsonBytes)).Err()
	if err != nil {
		http.Error(w, "Failed to publish task", http.StatusInternalServerError)
		return
	}

	// Store initial task state
	task := model.Task{
		ID:            taskCommand.ID,
		CodeLocation:  taskCommand.CodeLocation,
		TestsLocation: taskCommand.TestsLocation,
		Compiler:      taskCommand.Compiler,
		State:         model.CompilingTaskState,
	}
	taskJSON, err := json.Marshal(task)
	if err != nil {
		http.Error(w, "Failed to marshal task", http.StatusInternalServerError)
		return
	}
	h.redisClient.Set(ctx, fmt.Sprintf("task:%s", task.ID), string(taskJSON), 0)

	// Respond with task ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"task_id": task.ID})
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	taskID := vars["id"]

	taskJSON, err := h.redisClient.Get(ctx, fmt.Sprintf("task:%s", taskID)).Result()
	if err == redis.Nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve task", http.StatusInternalServerError)
		return
	}

	var task model.Task
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		http.Error(w, "Failed to unmarshal task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	keys, err := h.redisClient.Keys(ctx, "task:*").Result()
	if err != nil {
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	tasks := make([]model.Task, 0, len(keys))
	for _, key := range keys {
		taskJSON, err := h.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		var task model.Task
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
