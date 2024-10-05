Polling API Project
This is a simple polling API that allows authenticated users to create, vote, and fetch polls.
The project uses JWT for authentication and supports role-based access control for admin,
super-admin, and regular users.
Users can be enabled or disabled by admins or super-admins, and only super-admins can list all
users.
## Features
- User authentication with JWT
- Role-based access control (admin, super-admin, user)
- Poll creation and voting
- Soft delete for users (enable/disable users)
- SQLite database for user and poll data
- Middleware for authentication and role checking
## Project Structure
polling-api/
cmd/
 server/
 main.go
internal/
 handlers/
 models/
 database/
pkg/
 middleware/
 jwtutil/
go.mod
go.sum
README.md
.env
## Requirements
- Go 1.20+
- SQLite
- github.com/golang-jwt/jwt/v4 for JWT
## Setup
1. Clone the repository:
git clone https://github.com/your-repo/polling-api.git
cd polling-api
2. Install dependencies:
go mod download
3. Set up the .env file with the SQLite database path:
DB_PATH=./polls.db
4. Run the application:
go run cmd/server/main.go
## Test the API
### 1. Create Admin, Super-Admin, and Regular Users
curl -X POST http://localhost:8080/test
### 2. Login User
curl -X POST http://localhost:8080/login -d '{"username":"admin", "password":"adminpassword"}' -H
"Content-Type: application/json"
### 3. Logout User
curl -X POST http://localhost:8080/logout
### 4. Enable a User
curl -X POST "http://localhost:8080/users/enable?username=testuser" --cookie
"token=<admin-token>"
### 5. Disable a User
curl -X POST "http://localhost:8080/users/disable?username=testuser" --cookie
"token=<admin-token>"
### 6. Get All Polls
curl -X GET http://localhost:8080/polls/all --cookie "token=<user-token>"
## Example Response from /test
{
 "admin": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
 "superadmin": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
 "user": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
