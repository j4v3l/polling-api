package models

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Active   bool   `json:"active"`
    Role     string `json:"role"`    // New field to define user roles (e.g., "user", "admin", "super-admin")
}
