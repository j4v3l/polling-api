package jwtutil

import (
    "time"
    "github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your-secret-key")

// Claims struct to contain token claims, including the user's role
type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`  // Include role in JWT claims
    jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a user with their role
func GenerateJWT(username, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

// ValidateJWT validates a given JWT token and extracts claims
func ValidateJWT(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil || !token.Valid {
        return nil, err
    }

    return claims, nil
}
