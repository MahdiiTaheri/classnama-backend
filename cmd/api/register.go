package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

type ExecRegisterPayload struct {
	FirstName string `json:"first_name" validate:"required,max=72"`
	LastName  string `json:"last_name" validate:"required,max=72"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	Role      string `json:"role" validate:"required,oneof=admin manager"`
}

type TeacherRegisterPayload struct {
	FirstName   string `json:"first_name" validate:"required,max=72"`
	LastName    string `json:"last_name" validate:"required,max=72"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
	Subject     string `json:"subject" validate:"required,max=128"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	HireDate    string `json:"hire_date" validate:"required,datetime=2006-01-02"`
}

type StudentRegisterPayload struct {
	FirstName         string    `json:"first_name" validate:"required,max=72"`
	LastName          string    `json:"last_name" validate:"required,max=72"`
	Email             string    `json:"email" validate:"required,email"`
	Password          string    `json:"password" validate:"required,min=8,max=72"`
	PhoneNumber       *string   `json:"phone_number"`
	ClassRoomID       int64     `json:"classroom_id" validate:"required"`
	BirthDate         time.Time `json:"birth_date" validate:"required"`
	Address           string    `json:"address" validate:"required"`
	ParentName        string    `json:"parent_name" validate:"required"`
	ParentPhoneNumber string    `json:"parent_phone_number" validate:"required"`
	TeacherID         int64     `json:"teacher_id" validate:"required"`
}

// registerExecHandler godoc
//
//	@Summary		Register a new Exec
//	@Description	Only Execs with manager/admin roles can create new Execs
//	@Tags			Execs
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		ExecRegisterPayload	true	"Exec registration payload"
//	@Success		201		{object}	map[string]any		"Returns the created Exec and JWT token"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/execs/register [post]
func (app *application) registerExecHandler(w http.ResponseWriter, r *http.Request) {
	var payload ExecRegisterPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exec := &store.Exec{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Role:      store.Role(payload.Role),
	}
	if err := exec.Password.Set(payload.Password); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.store.Execs.Create(r.Context(), exec); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.createAndRespondJWT(w, r, exec, string(exec.Role))
}

// registerTeacherHandler godoc
//
//	@Summary		Register a new Teacher
//	@Description	Only Execs with manager/admin roles can create new Teachers
//	@Tags			Teachers
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		TeacherRegisterPayload	true	"Teacher registration payload"
//	@Success		201		{object}	store.Teacher			"Returns the created Teacher"
//	@Failure		400		{object}	map[string]string		"Bad request"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/teachers [post]
func (app *application) registerTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var payload TeacherRegisterPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	teacher := &store.Teacher{
		FirstName:   payload.FirstName,
		LastName:    payload.LastName,
		Email:       payload.Email,
		Subject:     payload.Subject,
		PhoneNumber: payload.PhoneNumber,
	}
	if err := teacher.Password.Set(payload.Password); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.store.Teachers.Create(r.Context(), teacher); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, teacher)
}

// registerStudentHandler godoc
//
//	@Summary		Register a new Student
//	@Description	Only Execs with manager/admin roles can create new Students
//	@Tags			Students
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		StudentRegisterPayload	true	"Student registration payload"
//	@Success		201		{object}	store.Student			"Returns the created Student"
//	@Failure		400		{object}	map[string]string		"Bad request"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/students [post]
func (app *application) registerStudentHandler(w http.ResponseWriter, r *http.Request) {
	var payload StudentRegisterPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	student := &store.Student{
		FirstName:         payload.FirstName,
		LastName:          payload.LastName,
		Email:             payload.Email,
		PhoneNumber:       payload.PhoneNumber,
		ClassRoomID:       payload.ClassRoomID,
		BirthDate:         payload.BirthDate,
		Address:           payload.Address,
		ParentName:        payload.ParentName,
		ParentPhoneNumber: payload.ParentPhoneNumber,
		TeacherID:         payload.TeacherID,
	}
	if err := student.Password.Set(payload.Password); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.store.Students.Create(r.Context(), student); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, student)
}

func (app *application) createAndRespondJWT(
	w http.ResponseWriter,
	r *http.Request,
	entity any,
	role string,
) {
	var id int64
	switch v := entity.(type) {
	case *store.Exec:
		id = v.ID
	case *store.Teacher:
		id = v.ID
	case *store.Student:
		id = v.ID
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

	if err := app.jsonResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}
