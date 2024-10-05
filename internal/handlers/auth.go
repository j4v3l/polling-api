
package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"
    "log"
    "strings"

    "polling-api/internal/database"
    "polling-api/internal/models"
    "polling-api/pkg/jwt"
    _ "github.com/mattn/go-sqlite3"
)

// Login handler for logging in users
func Login(w http.ResponseWriter, r *http.Request) {
    var credentials models.User
    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Query the user from the SQLite database
    query := `SELECT username, password, active, role FROM users WHERE username = ?`
    var user models.User
    err := database.DB.QueryRow(query, credentials.Username).Scan(&user.Username, &user.Password, &user.Active, &user.Role)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Check if the credentials are correct
    if credentials.Password != user.Password {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Check if the user is active
    if !user.Active {
        http.Error(w, "User account is disabled", http.StatusForbidden)
        return
    }

    // Generate JWT with the user's role
    token, err := jwtutil.GenerateJWT(user.Username, user.Role)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Set JWT in cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    token,
        Expires:  time.Now().Add(24 * time.Hour),
        Path:     "/",
        HttpOnly: true,
    })

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Login successful"))
}

// Logout handler for logging out users
func Logout(w http.ResponseWriter, r *http.Request) {
    // Clear the cookie by setting an expired date
    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    "",
        Expires:  time.Now().Add(-time.Hour),
        Path:     "/",
        HttpOnly: true,
    })

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Logout successful"))
}

// EnableUser handler to enable a user account (only for admins)
func EnableUser(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    if username == "" {
        http.Error(w, "Missing username parameter", http.StatusBadRequest)
        return
    }

    query := `UPDATE users SET active = 1 WHERE username = ?`
    _, err := database.DB.Exec(query, username)
    if err != nil {
        log.Printf("Error enabling user: %v", err)
        http.Error(w, "Error enabling user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("User enabled"))
}

// DisableUser handler to disable a user account (soft delete, only for admins)
func DisableUser(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    if username == "" {
        http.Error(w, "Missing username parameter", http.StatusBadRequest)
        return
    }

    query := `UPDATE users SET active = 0 WHERE username = ?`
    _, err := database.DB.Exec(query, username)
    if err != nil {
        log.Printf("Error disabling user: %v", err)
        http.Error(w, "Error disabling user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("User disabled"))
}

// ListUsers handler for listing all users (for super-admin)
func ListUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := database.DB.Query(`SELECT username, active, role FROM users`)
    if err != nil {
        log.Printf("Error querying users: %v", err)
        http.Error(w, "Error querying users", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        err := rows.Scan(&user.Username, &user.Active, &user.Role)
        if err != nil {
            log.Printf("Error scanning user: %v", err)
            http.Error(w, "Error scanning users", http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    // Return the list of users as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}


// TestRoute creates admin, super-admin, regular users, test polls, and votes for testing purposes
func TestRoute(w http.ResponseWriter, r *http.Request) {
    // Predefined users
    users := []models.User{
        {Username: "admin", Password: "adminpassword", Active: true, Role: "admin"},
        {Username: "superadmin", Password: "superpassword", Active: true, Role: "super-admin"},
        {Username: "user", Password: "userpassword", Active: true, Role: "user"},
    }

    // Insert users into the database
    for _, user := range users {
        query := `INSERT INTO users (username, password, active, role) VALUES (?, ?, ?, ?)`
        _, err := database.DB.Exec(query, user.Username, user.Password, user.Active, user.Role)
        if err != nil {
            log.Printf("Error inserting user %s: %v", user.Username, err)
            http.Error(w, "Error creating users", http.StatusInternalServerError)
            return
        }
    }

    // Generate JWT tokens for each user
    tokens := make(map[string]string)
    for _, user := range users {
        token, err := jwtutil.GenerateJWT(user.Username, user.Role)
        if err != nil {
            log.Printf("Error generating token for user %s: %v", user.Username, err)
            http.Error(w, "Error generating tokens", http.StatusInternalServerError)
            return
        }
        tokens[user.Username] = token
    }

    // Predefined polls
    polls := []models.Poll{
        {ID: "poll1", Question: "What's your favorite programming language?", Options: []string{"Go", "Python", "Rust"}, ExpiresAt: time.Now().AddDate(0, 0, -1)}, // Expired
        {ID: "poll2", Question: "What's your least favorite programming language?", Options: []string{"Go", "Python", "Rust"}, ExpiresAt: time.Now().AddDate(0, 0, 1)}, // Active
    }

    // Insert polls into the database
    for _, poll := range polls {
        optionsStr := strings.Join(poll.Options, ",")
        votesStr := strings.Repeat("0,", len(poll.Options))
        votesStr = strings.TrimSuffix(votesStr, ",") // remove trailing comma

        query := `INSERT INTO polls (id, question, options, votes, expires_at) VALUES (?, ?, ?, ?, ?)`
        _, err := database.DB.Exec(query, poll.ID, poll.Question, optionsStr, votesStr, poll.ExpiresAt.Format(time.RFC3339))
        if err != nil {
            log.Printf("Error inserting poll %s: %v", poll.ID, err)
            http.Error(w, "Error creating polls", http.StatusInternalServerError)
            return
        }
    }

    // Predefined votes for user
    votes := []struct {
        UserID  string
        PollID  string
        Option  string
    }{
        {"user", "poll1", "Go"},
        {"user", "poll2", "Python"},
    }

    // Insert votes into the database
    for _, vote := range votes {
        query := `INSERT INTO votes (user_id, poll_id, option, voted_at) VALUES (?, ?, ?, ?)`
        _, err := database.DB.Exec(query, vote.UserID, vote.PollID, vote.Option, time.Now())
        if err != nil {
            log.Printf("Error inserting vote for poll %s: %v", vote.PollID, err)
            http.Error(w, "Error creating votes", http.StatusInternalServerError)
            return
        }
    }

    // Return tokens and test data summary in the response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "tokens": tokens,
        "polls":  polls,
        "votes":  votes,
    })
}

