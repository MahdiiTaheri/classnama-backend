package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/auth"
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
		if !v.Password.Check(payload.Password) {
			app.unauthorizedResponse(w, r, fmt.Errorf("invalid credentials"))
			return
		}
		id = v.ID
		role = string(v.Role)
	case *store.Teacher:
		if !v.Password.Check(payload.Password) {
			app.unauthorizedResponse(w, r, fmt.Errorf("invalid credentials"))
			return
		}
		id = v.ID
		role = "teacher"
	case *store.Student:
		if !v.Password.Check(payload.Password) {
			app.unauthorizedResponse(w, r, fmt.Errorf("invalid credentials"))
			return
		}
		id = v.ID
		role = "student"
	default:
		app.internalServerErrorResponse(w, r, fmt.Errorf("unsupported entity type"))
		return
	}

	claims := &auth.Claims{
		ID:    id,
		Email: payload.Email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(id),
			Issuer:    app.config.auth.token.iss,
			Audience:  []string{app.config.auth.token.iss},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(app.config.auth.token.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
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

// loginExecHandler godoc
//
//	@Summary		Exec Login
//	@Description	Login as an Exec (admin or manager) and get a JWT token
//	@Tags			Execs
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginPayload		true	"Login payload"
//	@Success		200		{object}	map[string]any		"Returns the logged-in exec and JWT token"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Router			/execs/login [post]
func (app *application) loginExecHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		exec, err := app.store.Execs.GetByEmail(ctx, email)
		return exec, err
	})
}

// loginTeacherHandler godoc
//
//	@Summary		Teacher Login
//	@Description	Login as a Teacher and get a JWT token
//	@Tags			Teachers
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginPayload		true	"Login payload"
//	@Success		200		{object}	map[string]any		"Returns the logged-in teacher and JWT token"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Router			/teachers/login [post]
func (app *application) loginTeacherHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		teacher, err := app.store.Teachers.GetByEmail(ctx, email)
		return teacher, err
	})
}

// loginStudentHandler godoc
//
//	@Summary		Student Login
//	@Description	Login as a Student and get a JWT token
//	@Tags			Students
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginPayload		true	"Login payload"
//	@Success		200		{object}	map[string]any		"Returns the logged-in student and JWT token"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Router			/students/login [post]
func (app *application) loginStudentHandler(w http.ResponseWriter, r *http.Request) {
	app.loginHandler(w, r, func(ctx context.Context, email string) (any, error) {
		student, err := app.store.Students.GetByEmail(ctx, email)
		return student, err
	})
}
