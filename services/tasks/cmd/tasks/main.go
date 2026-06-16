package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"tasks/internal/cache"
	"tasks/internal/handlers"
	"tasks/internal/repository"
	"tasks/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse Redis configuration (standalone)
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	baseTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_SECONDS", "120"))
	jitter, _ := strconv.Atoi(getEnv("CACHE_TTL_JITTER_SECONDS", "30"))

	log.Printf("[INFO] Starting Tasks Service")
	log.Printf("[INFO] Redis addr: %s", redisAddr)
	log.Printf("[INFO] Cache config: TTL=%ds, Jitter=%ds", baseTTL, jitter)

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cache.CacheConfig{
		Addr:           redisAddr,
		Password:       getEnv("REDIS_PASSWORD", ""),
		BaseTTLSeconds: baseTTL,
		JitterSeconds:  jitter,
	})
	if err != nil {
		log.Printf("[WARN] Redis initialization warning: %v", err)
	}

	if redisCache != nil && redisCache.IsAvailable() {
		log.Printf("[INFO] Redis cache is available")
	} else {
		log.Printf("[WARN] Redis cache is NOT available - running in degraded mode")
	}

	// Initialize repository
	taskRepo := repository.NewInMemoryRepository()

	// Add demo data
	demoTask := &repository.Task{
		Title:       "Demo Task",
		Description: "This is a demo task for cache testing",
		DueDate:     "2026-01-20",
		Done:        false,
	}
	ctx := context.Background()
	if err := taskRepo.Create(ctx, demoTask); err != nil {
		log.Printf("[WARN] Failed to create demo task: %v", err)
	} else {
		log.Printf("[INFO] Created demo task with ID: %s", demoTask.ID)
	}

	// Initialize service with cache
	taskService := service.NewTaskService(taskRepo, redisCache)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks/", taskHandler.GetTask)
	mux.HandleFunc("GET /v1/tasks", taskHandler.GetTasks)
	mux.HandleFunc("POST /v1/tasks", taskHandler.CreateTask)
	mux.HandleFunc("PATCH /v1/tasks/", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /v1/tasks/", taskHandler.DeleteTask)

	port := getEnv("TASKS_PORT", "8082")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Printf("[INFO] Tasks service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}

	if redisCache != nil {
		redisCache.Close()
	}

	log.Println("[INFO] Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
