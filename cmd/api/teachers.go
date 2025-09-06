package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/MahdiiTaheri/classnama-backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type teacherKey string

const teacherCtx teacherKey = "teacher"

type CreateTeacherPayload struct {
	FirstName   string `json:"first_name" validate:"required,max=72"`
	LastName    string `json:"last_name" validate:"required,max=72"`
	Email       string `json:"email" validate:"required,email"`
	Subject     string `json:"subject" validate:"required,max=128"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	HireDate    string `json:"hire_date" validate:"required,datetime=2006-01-02"`
}

type UpdateTeacherPayload struct {
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,max=72"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,max=72"`
	Email       *string `json:"email,omitempty" validate:"omitempty,email"`
	Subject     *string `json:"subject,omitempty" validate:"omitempty,max=128"`
	PhoneNumber *string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	HireDate    *string `json:"hire_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// CreateTeacher godoc
//
//	@Summary		Create a new teacher
//	@Description	Creates a new teacher with first name, last name, email, subject, and optional phone.
//	@Tags			teachers
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateTeacherPayload	true	"Teacher payload"
//	@Success		201		{object}	store.Teacher			"Returns the created teacher"
//	@Failure		400		{object}	error					"Validation failed"
//	@Failure		409		{object}	error					"Conflict, teacher already exists"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/teachers [post]
//	@ID				createTeacher
func (app *application) createTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateTeacherPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
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
	ctx := r.Context()

	if err := app.store.Teachers.Create(ctx, teacher); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, teacher); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}
}

// GetTeachers godoc
//
//	@Summary	Get all teachers
//	@Tags		teachers
//	@Produce	json
//	@Success	200	{array}		store.Teacher
//	@Failure	500	{object}	error
//	@Security	ApiKeyAuth
//	@Router		/teachers [get]
//	@ID			getTeachers
func (app *application) getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teachers, err := app.store.Teachers.GetAll(ctx)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, teachers); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// GetTeacher godoc
//
//	@Summary	Get a teacher by ID
//	@Tags		teachers
//	@Produce	json
//	@Param		teacherID	path		int	true	"Teacher ID"
//	@Success	200			{object}	store.Teacher
//	@Failure	404			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/teachers/{teacherID} [get]
//	@ID			getTeacher
func (app *application) getTeacherHandler(w http.ResponseWriter, r *http.Request) {
	teacher := getTeacherFromCtx(r)
	if teacher == nil {
		app.notfoundResponse(w, r, fmt.Errorf("teacher not found in context"))
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, teacher); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// UpdateTeacher godoc
//
//	@Summary	Update a teacher
//	@Tags		teachers
//	@Accept		json
//	@Produce	json
//	@Param		teacherID	path		int						true	"Teacher ID"
//	@Param		payload		body		UpdateTeacherPayload	true	"Teacher update payload"
//	@Success	200			{object}	store.Teacher
//	@Failure	400			{object}	error
//	@Failure	404			{object}	error
//	@Failure	409			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/teachers/{teacherID} [patch]
//	@ID			updateTeacher
func (app *application) updateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	teacher := getTeacherFromCtx(r)
	if teacher == nil {
		app.notfoundResponse(w, r, fmt.Errorf("teacher not found"))
		return
	}

	var payload UpdateTeacherPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Apply non-nil fields using reflection
	utils.ApplyPatch(teacher, payload)

	// Update in DB
	if err := app.store.Teachers.Update(r.Context(), teacher); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		case store.ErrNotFound:
			app.notfoundResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	// Return updated teacher
	if err := app.jsonResponse(w, http.StatusOK, teacher); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// DeleteTeacher godoc
//
//	@Summary	Delete a teacher
//	@Tags		teachers
//	@Param		teacherID	path	int	true	"Teacher ID"
//	@Success	204			"No Content"
//	@Failure	404			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/teachers/{teacherID} [delete]
//	@ID			deleteTeacher
func (app *application) deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "teacherID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Teachers.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notfoundResponse(w, r, err)
		default:
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Middleware ---

func (app *application) teachersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "teacherID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid teacher ID"))
			return
		}

		teacher, err := app.store.Teachers.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				app.notfoundResponse(w, r, err)
				return
			}
			app.internalServerErrorResponse(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), teacherCtx, teacher)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTeacherFromCtx(r *http.Request) *store.Teacher {
	teacher, _ := r.Context().Value(teacherCtx).(*store.Teacher)
	return teacher
}
