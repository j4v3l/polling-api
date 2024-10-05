
package handlers

import (
    "encoding/json"
    "net/http"
    "strings"
    "sync"
    "log"
    "time"
    "database/sql"
    "polling-api/internal/models"
    "polling-api/internal/database"
    "strconv"
)

var polls = make(map[string]models.Poll)
var mu sync.Mutex

func CreatePoll(w http.ResponseWriter, r *http.Request) {
    var poll models.Poll
    if err := json.NewDecoder(r.Body).Decode(&poll); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Convert options and votes to comma-separated strings
    optionsStr := strings.Join(poll.Options, ",")
    votesStr := strings.Repeat("0,", len(poll.Options))
    votesStr = strings.TrimSuffix(votesStr, ",") // remove trailing comma

    // Insert the poll into the SQLite database
    query := `INSERT INTO polls (id, question, options, votes, expires_at) VALUES (?, ?, ?, ?, ?)`
    _, err := database.DB.Exec(query, poll.ID, poll.Question, optionsStr, votesStr, poll.ExpiresAt.Format(time.RFC3339))
    if err != nil {
        http.Error(w, "Error inserting poll into database", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(poll)
}



func GetPoll(w http.ResponseWriter, r *http.Request) {
    pollID := r.URL.Query().Get("id")

    // Fetch the poll from the database
    query := `SELECT question, options, votes, expires_at FROM polls WHERE id = ?`
    var question, optionsStr, votesStr, expiresAtStr string
    err := database.DB.QueryRow(query, pollID).Scan(&question, &optionsStr, &votesStr, &expiresAtStr)
    if err == sql.ErrNoRows {
        http.Error(w, "Poll not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Error fetching poll from database", http.StatusInternalServerError)
        return
    }

    // Convert options and votes from comma-separated strings to slices
    options := strings.Split(optionsStr, ",")
    votes := strings.Split(votesStr, ",")
    expiresAt, _ := time.Parse(time.RFC3339, expiresAtStr)

    // Create a poll object to return
    poll := models.Poll{
        ID:        pollID,
        Question:  question,
        Options:   options,
        Votes:     make([]int, len(votes)),
        ExpiresAt: expiresAt,
    }

    // Convert votes from string to int
    for i, voteStr := range votes {
        poll.Votes[i], _ = strconv.Atoi(voteStr)
    }

    // Send poll as JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(poll)
}

func GetAllPolls(w http.ResponseWriter, r *http.Request) {
    // Query all polls from the database
    query := `SELECT id, question, options, votes, expires_at FROM polls`
    rows, err := database.DB.Query(query)
    if err != nil {
        http.Error(w, "Error fetching polls from database", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var polls []models.Poll

    // Loop through the rows and append each poll to the polls slice
    for rows.Next() {
        var poll models.Poll
        var optionsStr, votesStr, expiresAtStr string
        err := rows.Scan(&poll.ID, &poll.Question, &optionsStr, &votesStr, &expiresAtStr)
        if err != nil {
            http.Error(w, "Error scanning poll from database", http.StatusInternalServerError)
            return
        }

        // Convert options and votes from comma-separated strings to slices
        poll.Options = strings.Split(optionsStr, ",")
        votes := strings.Split(votesStr, ",")
        poll.Votes = make([]int, len(votes))
        for i, voteStr := range votes {
            poll.Votes[i], _ = strconv.Atoi(voteStr)
        }

        // Parse the expiration date
        poll.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)

        polls = append(polls, poll)
    }

    if err = rows.Err(); err != nil {
        http.Error(w, "Error iterating through polls", http.StatusInternalServerError)
        return
    }

    // Return the polls as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(polls)
}

func UpdatePoll(w http.ResponseWriter, r *http.Request) {
    var poll models.Poll
    if err := json.NewDecoder(r.Body).Decode(&poll); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    query := `UPDATE polls SET question = ?, options = ?, expires_at = ? WHERE id = ?`
    _, err := database.DB.Exec(query, poll.Question, poll.Options, poll.ExpiresAt, poll.ID)
    if err != nil {
        log.Printf("Error updating poll: %v", err)
        http.Error(w, "Error updating poll", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Poll updated"))
}

func DeletePoll(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing poll ID", http.StatusBadRequest)
        return
    }

    query := `DELETE FROM polls WHERE id = ?`
    _, err := database.DB.Exec(query, id)
    if err != nil {
        log.Printf("Error deleting poll: %v", err)
        http.Error(w, "Error deleting poll", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Poll deleted"))
}
// SummarizePollResults checks for expired polls and summarizes their results

func SummarizePollResults() {
    log.Println("Checking for expired polls to summarize...")

    // Fetch polls that have expired and are not summarized yet
    query := `SELECT id, options, votes FROM polls WHERE expires_at < CURRENT_TIMESTAMP AND id NOT IN (SELECT poll_id FROM poll_summary)`
    rows, err := database.DB.Query(query)
    if err != nil {
        log.Printf("Error fetching expired polls: %v", err)
        return
    }
    defer rows.Close()

    // Start a transaction
    tx, err := database.DB.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        return
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback() // Rollback the transaction in case of panic
            log.Printf("Transaction rolled back due to panic: %v", p)
        }
    }()

    // Iterate through expired polls
    for rows.Next() {
        var pollID, optionsStr, votesStr string
        err := rows.Scan(&pollID, &optionsStr, &votesStr)
        if err != nil {
            log.Printf("Error scanning poll: %v", err)
            continue
        }

        log.Printf("Summarizing poll: %s", pollID)

        // Parse options and votes
        options := strings.Split(optionsStr, ",")
        votes := strings.Split(votesStr, ",")

        // Calculate total votes and determine the winning option
        totalVotes := 0
        winningOption := ""
        maxVotes := -1
        for i, voteStr := range votes {
            voteCount := parseVoteCount(voteStr)
            totalVotes += voteCount
            if voteCount > maxVotes {
                maxVotes = voteCount
                winningOption = options[i]
            }
        }

        log.Printf("Poll %s summary - Total votes: %d, Winning option: %s", pollID, totalVotes, winningOption)

        // Store the summary in the poll_summary table using the transaction
        if err := storePollSummary(tx, pollID, totalVotes, winningOption); err != nil {
            log.Printf("Error storing poll summary for poll %s: %v", pollID, err)
            tx.Rollback()  // Rollback the transaction if there's an error
            return
        }
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error iterating through polls: %v", err)
    }

    // Commit the transaction if no errors
    if err := tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
    } else {
        log.Println("Poll summarization transaction committed successfully.")
    }
}


// Helper function to parse vote count safely
func parseVoteCount(voteStr string) int {
    voteCount, err := strconv.Atoi(voteStr)
    if err != nil {
        return 0
    }
    return voteCount
}


// Helper function to store poll summary in the database using a transaction
func storePollSummary(tx *sql.Tx, pollID string, totalVotes int, winningOption string) error {
    query := `INSERT INTO poll_summary (poll_id, total_votes, winning_option) VALUES (?, ?, ?)`
    _, err := tx.Exec(query, pollID, totalVotes, winningOption)
    if err != nil {
        log.Printf("Error inserting poll summary: %v", err)
        return err
    }
    log.Printf("Poll summary inserted for poll_id: %s", pollID)
    return nil
}

// GetPollSummary returns the summary of a given poll
func GetPollSummary(w http.ResponseWriter, r *http.Request) {
    pollID := r.URL.Query().Get("poll_id")
    if pollID == "" {
        http.Error(w, "Missing poll_id parameter", http.StatusBadRequest)
        return
    }

    query := `SELECT total_votes, winning_option, summary_time FROM poll_summary WHERE poll_id = ?`
    var totalVotes int
    var winningOption string
    var summaryTime string

    err := database.DB.QueryRow(query, pollID).Scan(&totalVotes, &winningOption, &summaryTime)
    if err == sql.ErrNoRows {
        http.Error(w, "No summary found for the given poll", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Error fetching poll summary", http.StatusInternalServerError)
        return
    }

    // Return the summary in JSON format
    json.NewEncoder(w).Encode(map[string]interface{}{
        "poll_id":       pollID,
        "total_votes":   totalVotes,
        "winning_option": winningOption,
        "summary_time":  summaryTime,
    })
}

func TriggerPollSummary(w http.ResponseWriter, r *http.Request) {
    SummarizePollResults()  // Manually trigger poll summarization
    w.Write([]byte("Poll summary triggered"))
}
