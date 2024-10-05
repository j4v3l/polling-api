package middleware

import (
    "net/http"
    "context"
    "polling-api/pkg/jwt"
)

// AuthMiddleware validates the JWT token for general authenticated users
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the JWT token from the cookies
        tokenCookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Validate the JWT token
        claims, err := jwtutil.ValidateJWT(tokenCookie.Value)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Optionally add claims (like user info) to the request context
        ctx := context.WithValue(r.Context(), "userID", claims.Username)
        ctx = context.WithValue(ctx, "userRole", claims.Role)

        // Pass the request to the next handler with the updated context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
// AdminMiddleware allows only users with "admin" or "super-admin" roles
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the JWT token from the cookies
        tokenCookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Validate the JWT token
        claims, err := jwtutil.ValidateJWT(tokenCookie.Value)
        if err != nil || (claims.Role != "admin" && claims.Role != "super-admin") {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        // Pass the request to the next handler if the role is valid
        next.ServeHTTP(w, r)
    })
}

// SuperAdminMiddleware allows only users with the "super-admin" role
func SuperAdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the JWT token from the cookies
        tokenCookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Validate the JWT token
        claims, err := jwtutil.ValidateJWT(tokenCookie.Value)
        if err != nil || claims.Role != "super-admin" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        // Pass the request to the next handler if the role is valid
        next.ServeHTTP(w, r)
    })
}
