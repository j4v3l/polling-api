
package database

import (
    "database/sql"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database and sets up the schema
func InitDB() {
    dbPath := os.Getenv("DB_PATH")
    if dbPath == "" {
        log.Fatalf("DB_PATH environment variable is not set")
    }

    var err error
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    // Create Polls table
    createPollsTableQuery := `CREATE TABLE IF NOT EXISTS polls (
        id TEXT PRIMARY KEY,
        question TEXT,
        options TEXT,
        votes TEXT,
        expires_at DATETIME
    );`
    
    _, err = DB.Exec(createPollsTableQuery)
    if err != nil {
        log.Fatalf("Error creating polls table: %v", err)
    }

    // Create Users table
    createUsersTableQuery := `CREATE TABLE IF NOT EXISTS users (
        username TEXT PRIMARY KEY,
        password TEXT NOT NULL,
        active INTEGER NOT NULL DEFAULT 1,  -- 1 for active, 0 for disabled
        role TEXT NOT NULL                  -- user, admin, or super-admin
    );`

    _, err = DB.Exec(createUsersTableQuery)
    if err != nil {
        log.Fatalf("Error creating users table: %v", err)
    }

    // Create Votes table
    createVotesTableQuery := `CREATE TABLE IF NOT EXISTS votes (
        user_id TEXT,
        poll_id TEXT,
        option TEXT,
        voted_at DATETIME,
        FOREIGN KEY (user_id) REFERENCES users(username),
        FOREIGN KEY (poll_id) REFERENCES polls(id)
    );`

    _, err = DB.Exec(createVotesTableQuery)
    if err != nil {
        log.Fatalf("Error creating votes table: %v", err)
    }
        // Create Poll Summary table (for expired polls)
    createPollSummaryTableQuery := `CREATE TABLE IF NOT EXISTS poll_summary (
        poll_id TEXT PRIMARY KEY,
        total_votes INTEGER,
        winning_option TEXT,
        summary_time DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (poll_id) REFERENCES polls(id)
    );`

    _, err = DB.Exec(createPollSummaryTableQuery)
    if err != nil {
        log.Fatalf("Error creating poll summary table: %v", err)
    }
}
