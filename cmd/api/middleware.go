package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	ID   int64
	Role string
	Data any
}

type contextKey string

const authUserCtx contextKey = "authUser"

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		jwtToken, err := app.authenticator.ValidateToken(parts[1])
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			app.unauthorizedResponse(w, r, fmt.Errorf("invalid claims"))
			return
		}

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			app.unauthorizedResponse(w, r, fmt.Errorf("role claim missing"))
			return
		}

		ctx := r.Context()
		var authUser AuthUser

		switch role {
		case "admin":
			user, err := app.store.Execs.GetByID(ctx, userID)
			if err != nil {
				app.unauthorizedResponse(w, r, err)
				return
			}
			authUser = AuthUser{ID: user.ID, Role: "admin", Data: user}

		case "manager":
			user, err := app.store.Execs.GetByID(ctx, userID)
			if err != nil {
				app.unauthorizedResponse(w, r, err)
				return
			}
			authUser = AuthUser{ID: user.ID, Role: "admin", Data: user}

		case "student":
			user, err := app.store.Students.GetByID(ctx, userID)
			if err != nil {
				app.unauthorizedResponse(w, r, err)
				return
			}
			authUser = AuthUser{ID: user.ID, Role: "student", Data: user}

		case "teacher":
			user, err := app.store.Teachers.GetByID(ctx, userID)
			if err != nil {
				app.unauthorizedResponse(w, r, err)
				return
			}
			authUser = AuthUser{ID: user.ID, Role: "teacher", Data: user}

		default:
			app.unauthorizedResponse(w, r, fmt.Errorf("unknown role"))
			return
		}

		ctx = context.WithValue(ctx, authUserCtx, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) requireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{})
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authUser, ok := r.Context().Value(authUserCtx).(AuthUser)
			if !ok {
				app.unauthorizedResponse(w, r, fmt.Errorf("user not found in context"))
				return
			}

			if _, allowed := roleSet[authUser.Role]; !allowed {
				app.forbiddenResponse(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
