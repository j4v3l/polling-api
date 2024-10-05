
package main

import (
    "log"
    "net/http"
    "polling-api/internal/database"
    "polling-api/pkg/middleware"
    "github.com/joho/godotenv"
    "polling-api/internal/handlers"
    "time"
    "os"
)

func main() {
    // Load environment variables from .env file (if exists)
    if _, err := os.Stat(".env"); err == nil {
        err := godotenv.Load()
        if err != nil {
            log.Fatalf("Error loading .env file")
        }
    }

    // Initialize the SQLite database
    database.InitDB()
    
    // Start background poll summarization goroutine
    go startAutoSummarization()

    mux := http.NewServeMux()

    // Vote-related routes for authenticated users 
    mux.Handle("/vote", middleware.AuthMiddleware(http.HandlerFunc(handlers.VotePoll)))
    mux.Handle("/vote/history", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetVoteHistory)))

    // Public routes
    mux.HandleFunc("/test", handlers.TestRoute)  // Test route to create users and tokens
    mux.HandleFunc("/polls/summarize", handlers.TriggerPollSummary)
    mux.HandleFunc("/login", handlers.Login)
    mux.HandleFunc("/logout", handlers.Logout)
    mux.HandleFunc("/poll/summary", handlers.GetPollSummary)

    // Poll-related routes for authenticated users
    mux.Handle("/polls", middleware.AuthMiddleware(http.HandlerFunc(handlers.CreatePoll)))
    mux.Handle("/polls/vote", middleware.AuthMiddleware(http.HandlerFunc(handlers.VotePoll)))
    mux.Handle("/polls/get", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetPoll)))
    mux.Handle("/polls/all", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetAllPolls)))
    mux.Handle("/polls/create", middleware.AdminMiddleware(http.HandlerFunc(handlers.CreatePoll)))
    mux.Handle("/polls/update", middleware.AdminMiddleware(http.HandlerFunc(handlers.UpdatePoll)))
    mux.Handle("/polls/delete", middleware.SuperAdminMiddleware(http.HandlerFunc(handlers.DeletePoll)))

    // Admin routes (only "admin" and "super-admin" can access)
    mux.Handle("/users/enable", middleware.AdminMiddleware(http.HandlerFunc(handlers.EnableUser)))
    mux.Handle("/users/disable", middleware.AdminMiddleware(http.HandlerFunc(handlers.DisableUser)))

    // Super-admin routes (only "super-admin" can access)
    mux.Handle("/admin/users", middleware.SuperAdminMiddleware(http.HandlerFunc(handlers.ListUsers)))

    // Apply logging middleware
    loggedMux := middleware.Logging(mux)

    log.Println("Server running on :8080")
    http.ListenAndServe(":8080", loggedMux)
}
// startAutoSummarization starts a goroutine that automatically summarizes polls every minute
func startAutoSummarization() {
    ticker := time.NewTicker(1 * time.Minute) // Runs every 1 minute
    defer ticker.Stop()

    for {
        <-ticker.C
        log.Println("Running automatic poll summarization...")
        handlers.SummarizePollResults() // Trigger the summarization
    }
}
