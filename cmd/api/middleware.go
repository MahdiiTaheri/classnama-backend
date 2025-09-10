package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MahdiiTaheri/classnama-backend/internal/auth"
)

type AuthUser struct {
	ID   int64
	Role string
	Data any
}

type ctxKey string

const userCtxKey ctxKey = "user"

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := app.authenticator.ValidateToken(tokenStr)
		if err != nil || token == nil || !token.Valid {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		claims, ok := token.Claims.(*auth.Claims)
		if !ok || claims == nil {
			app.unauthorizedResponse(w, r, fmt.Errorf("invalid token claims"))
			return
		}

		// put claims in context
		ctx := context.WithValue(r.Context(), userCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := getUser(r)
			if claims == nil {
				app.unauthorizedResponse(w, r, fmt.Errorf("missing claims"))
				return
			}

			if _, ok := allowed[claims.Role]; !ok {
				app.forbiddenResponse(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getUser(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(userCtxKey).(*auth.Claims)
	return claims
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.ratelimiter.Enabled {
			if allow, retryAfter := app.ratelimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
