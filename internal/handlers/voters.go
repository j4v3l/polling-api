package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "database/sql"
    "strings"
    "strconv"
    "polling-api/internal/database"
    "polling-api/internal/models"
    "time"
)



func VotePoll(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    pollID := r.URL.Query().Get("id")
    option := r.URL.Query().Get("option")

    // Check if user already voted on this poll
    existingVoteQuery := `SELECT COUNT(*) FROM votes WHERE user_id = ? AND poll_id = ?`
    var count int
    err := database.DB.QueryRow(existingVoteQuery, userID, pollID).Scan(&count)
    if err != nil {
        log.Printf("Error checking existing vote: %v", err)
        http.Error(w, "Error processing vote", http.StatusInternalServerError)
        return
    }
    if count > 0 {
        http.Error(w, "User has already voted on this poll", http.StatusForbidden)
        return
    }

    // Validate if the selected option is valid
    pollQuery := `SELECT options, votes FROM polls WHERE id = ?`
    var optionsStr, votesStr string
    err = database.DB.QueryRow(pollQuery, pollID).Scan(&optionsStr, &votesStr)
    if err == sql.ErrNoRows {
        http.Error(w, "Poll not found", http.StatusNotFound)
        return
    } else if err != nil {
        log.Printf("Error fetching poll options: %v", err)
        http.Error(w, "Error processing vote", http.StatusInternalServerError)
        return
    }

    options := strings.Split(optionsStr, ",")
    votes := strings.Split(votesStr, ",")
    if !isValidOption(option, options) {
        http.Error(w, "Invalid poll option", http.StatusBadRequest)
        return
    }

    // Find the index of the selected option
    optionIndex := -1
    for i, opt := range options {
        if opt == option {
            optionIndex = i
            break
        }
    }
    if optionIndex == -1 {
        http.Error(w, "Invalid option selected", http.StatusBadRequest)
        return
    }

    // Increment the vote count for the selected option
    currentVoteCount, _ := strconv.Atoi(votes[optionIndex])
    votes[optionIndex] = strconv.Itoa(currentVoteCount + 1)

    // Update the votes field in the polls table
    newVotesStr := strings.Join(votes, ",")
    updatePollVotesQuery := `UPDATE polls SET votes = ? WHERE id = ?`
    _, err = database.DB.Exec(updatePollVotesQuery, newVotesStr, pollID)
    if err != nil {
        log.Printf("Error updating poll votes: %v", err)
        http.Error(w, "Error recording vote", http.StatusInternalServerError)
        return
    }

    // Insert the vote into the votes table
    query := `INSERT INTO votes (user_id, poll_id, option, voted_at) VALUES (?, ?, ?, ?)`
    _, err = database.DB.Exec(query, userID, pollID, option, time.Now())
    if err != nil {
        log.Printf("Error recording vote: %v", err)
        http.Error(w, "Error recording vote", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Vote recorded"))
}

// Helper function to validate poll options
func isValidOption(option string, options []string) bool {
    for _, validOption := range options {
        if option == validOption {
            return true
        }
    }
    return false
}

// GetVoteHistory: Regular users can view their voting history
func GetVoteHistory(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)

    query := `SELECT poll_id, option, voted_at FROM votes WHERE user_id = ? ORDER BY voted_at DESC`
    rows, err := database.DB.Query(query, userID)
    if err != nil {
        log.Printf("Error querying vote history: %v", err)
        http.Error(w, "Error querying vote history", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var votes []models.Vote
    for rows.Next() {
        var vote models.Vote
        err := rows.Scan(&vote.PollID, &vote.Option, &vote.VotedAt)
        if err != nil {
            log.Printf("Error scanning vote history: %v", err)
            http.Error(w, "Error reading vote history", http.StatusInternalServerError)
            return
        }
        votes = append(votes, vote)
    }

    if len(votes) == 0 {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    json.NewEncoder(w).Encode(votes)
}

