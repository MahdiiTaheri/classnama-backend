package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/MahdiiTaheri/classnama-backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type studentKey string

const studentCtx studentKey = "student"

type CreateStudentPayload struct {
	FirstName         string `json:"first_name" validate:"required,max=72"`
	LastName          string `json:"last_name" validate:"required,max=72"`
	Email             string `json:"email" validate:"required,email"`
	PhoneNumber       string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	Class             string `json:"class" validate:"required,max=16"`
	BirthDate         string `json:"birth_date" validate:"required,datetime=2006-01-02"`
	Address           string `json:"address" validate:"required,max=256"`
	ParentName        string `json:"parent_name" validate:"required,max=128"`
	TeacherID         int64  `json:"teacher_id" validate:"required"`
	ParentPhoneNumber string `json:"parent_phone_number" validate:"required,e164"`
}

type UpdateStudentPayload struct {
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,max=72"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,max=72"`
	Email             *string `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumber       *string `json:"phone_number,omitempty" validate:"omitempty,e164"`
	Class             *string `json:"class,omitempty" validate:"omitempty,max=16"`
	BirthDate         *string `json:"birth_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Address           *string `json:"address,omitempty" validate:"omitempty,max=256"`
	ParentName        *string `json:"parent_name,omitempty" validate:"omitempty,max=128"`
	ParentPhoneNumber *string `json:"parent_phone_number,omitempty" validate:"omitempty,e164"`
	TeacherID         *int64  `json:"teacher_id,omitempty" validate:"omitempty"`
}

// CreateStudent godoc
//
//	@Summary		Create a new student
//	@Description	Creates a new student with first name, last name, email, subject, and optional phone.
//	@Tags			students
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateStudentPayload	true	"student payload"
//	@Success		201		{object}	store.student			"Returns the created student"
//	@Failure		400		{object}	error					"Validation failed"
//	@Failure		409		{object}	error					"Conflict, student already exists"
//	@Failure		500		{object}	error					"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/students [post]
//	@ID				createstudent
func (app *application) createStudentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateStudentPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	birthDate, err := time.Parse("2006-01-02", payload.BirthDate)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid birth_date format: %w", err))
		return
	}

	student := &store.Student{
		FirstName:         payload.FirstName,
		LastName:          payload.LastName,
		Email:             payload.Email,
		PhoneNumber:       &payload.PhoneNumber,
		Class:             payload.Class,
		BirthDate:         birthDate,
		Address:           payload.Address,
		ParentName:        payload.ParentName,
		ParentPhoneNumber: payload.ParentPhoneNumber,
		TeacherID:         payload.TeacherID,
	}

	ctx := r.Context()

	if err := app.store.Students.Create(ctx, student); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, student); err != nil {
		switch err {
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}
}

// GetStudents godoc
//
//	@Summary	Get all students
//	@Tags		students
//	@Produce	json
//	@Success	200	{array}		store.student
//	@Failure	500	{object}	error
//	@Security	ApiKeyAuth
//	@Router		/students [get]
//	@ID			getStudents
func (app *application) getStudentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	students, err := app.store.Students.GetAll(ctx)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, students); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// Getstudent godoc
//
//	@Summary	Get a student by ID
//	@Tags		students
//	@Produce	json
//	@Param		studentID	path		int	true	"student ID"
//	@Success	200			{object}	store.student
//	@Failure	404			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/students/{studentID} [get]
//	@ID			getstudent
func (app *application) getStudentHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)
	if student == nil {
		app.notfoundResponse(w, r, fmt.Errorf("student not found in context"))
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, student); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// UpdateStudent godoc
//
//	@Summary	Update a student
//	@Tags		students
//	@Accept		json
//	@Produce	json
//	@Param		studentID	path		int						true	"student ID"
//	@Param		payload		body		UpdateStudentPayload	true	"student update payload"
//	@Success	200			{object}	store.student
//	@Failure	400			{object}	error
//	@Failure	404			{object}	error
//	@Failure	409			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/students/{studentID} [patch]
//	@ID			updateStudent
func (app *application) updateStudentHandler(w http.ResponseWriter, r *http.Request) {
	student := getStudentFromCtx(r)
	if student == nil {
		app.notfoundResponse(w, r, fmt.Errorf("student not found"))
		return
	}

	var payload UpdateStudentPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Apply non-nil fields using reflection
	utils.ApplyPatch(student, payload)

	// Update in DB
	if err := app.store.Students.Update(r.Context(), student); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notfoundResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	// Return updated student
	if err := app.jsonResponse(w, http.StatusOK, student); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// DeleteStudent godoc
//
//	@Summary	Delete a student
//	@Tags		students
//	@Param		studentID	path	int	true	"student ID"
//	@Success	204			"No Content"
//	@Failure	404			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/students/{studentID} [delete]
//	@ID			deleteStudent
func (app *application) deleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "studentID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Students.Delete(ctx, id); err != nil {
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

func (app *application) studentsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "studentID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid student ID"))
			return
		}

		student, err := app.store.Students.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				app.notfoundResponse(w, r, err)
				return
			}
			app.internalServerErrorResponse(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), studentCtx, student)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getStudentFromCtx(r *http.Request) *store.Student {
	student, _ := r.Context().Value(studentCtx).(*store.Student)
	return student
}
