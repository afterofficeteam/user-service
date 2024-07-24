package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"user-service/src/util/helper/jwt"
)

const userKey = "UserID"
const roleKey = "Role"

const (
	RoleAdmin  = "Admin"
	RoleUser   = "User"
	RoleSeller = "Seller"
)

func SetUserID(ctx context.Context, userID string) context.Context {
	ctx = context.WithValue(ctx, userKey, userID)
	return ctx
}

func SetRole(ctx context.Context, role string) context.Context {
	ctx = context.WithValue(ctx, roleKey, role)
	return ctx
}

func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(userKey).(string)
	if !ok {
		return ""
	}

	return userID
}

func GetRole(ctx context.Context) string {
	role, ok := ctx.Value(roleKey).(string)
	if !ok {
		return ""
	}

	return role
}

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Message": "Unauthorized",
				"Data":    nil,
			})
			return
		}

		tokenString = tokenString[len("Bearer "):]
		payload, err := jwt.VerifyToken(tokenString)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Message": "Unauthorized",
				"Data":    nil,
			})
			return
		}

		ctx = SetUserID(ctx, payload.UserID)
		ctx = SetRole(ctx, payload.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
