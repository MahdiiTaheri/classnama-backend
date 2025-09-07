package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type RegisterPayload struct {
	FirstName string `json:"first_name" validate:"required,max=72"`
	LastName  string `json:"last_name" validate:"required,max=72"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	Role      string `json:"role,omitempty" validate:"omitempty,oneof=admin manager"`
}

func (app *application) loginHandler(
	w http.ResponseWriter,
	r *http.Request,
	getByEmail func(ctx context.Context, email string) (any, error)) {
	var payload LoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	entity, err := getByEmail(ctx, payload.Email)
	if err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}

	var id int64
	var role string

	switch v := entity.(type) {
	case *store.Exec:
		id = v.ID
		role = string(v.Role) // Exec has a role column
	case *store.Teacher:
		id = v.ID
		role = "teacher" // fixed role for JWT
	case *store.Student:
		id = v.ID
		role = "student" // fixed role for JWT
	default:
		app.internalServerErrorResponse(w, r, fmt.Errorf("unsupported entity type"))
		return
	}

	claims := jwt.MapClaims{
		"sub":  id,
		"role": role,
		"exp":  time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat":  time.Now().Unix(),
		"nbf":  time.Now().Unix(),
		"iss":  app.config.auth.token.iss,
		"aud":  app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	resp := map[string]any{
		"entity": entity,
		"token":  token,
	}

	if err := app.jsonResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

func (app *application) loginExecHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		exec, err := app.store.Execs.GetByEmail(ctx, email)
		return exec, err
	})
}

func (app *application) loginTeacherHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		teacher, err := app.store.Teachers.GetByEmail(ctx, email)
		return teacher, err
	})
}

func (app *application) loginStudentHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		student, err := app.store.Students.GetByEmail(ctx, email)
		return student, err
	})
}
