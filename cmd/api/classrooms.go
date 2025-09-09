package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/MahdiiTaheri/classnama-backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type ClassroomRegisterPayload struct {
	Name     string `json:"name" validate:"required,max=128"`
	Capacity int64  `json:"capacity" validate:"required,min=1"`
	Grade    int64  `json:"grade,omitempty" validate:"required,min=1"`
}

type UpdateClassroomPayload struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,max=128"`
	Capacity *int64  `json:"capacity,omitempty" validate:"omitempty,min=5,max=40"`
	Grade    *int64  `json:"grade,omitempty" validate:"omitempty,min=1,max=30"`
}

type classroomKey string

const classroomCtx classroomKey = "classroom"

func (app *application) registerClassroomHandler(w http.ResponseWriter, r *http.Request) {
	var payload ClassroomRegisterPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	classroom := &store.Classroom{
		Name:     payload.Name,
		Capacity: payload.Capacity,
		Grade:    payload.Grade,
	}

	if err := app.store.Classrooms.Create(r.Context(), classroom); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, classroom)
}

// getClassroomsHandler (paginated, searchable)
func (app *application) getClassroomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pq := store.PaginatedQuery{Limit: 10, Offset: 0, SortBy: "id", Order: "asc"}
	pq, err := pq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(pq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	classrooms, err := app.store.Classrooms.GetAll(ctx, pq)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, classrooms)
}

// getClassroomHandler
func (app *application) getClassroomHandler(w http.ResponseWriter, r *http.Request) {
	classroom := getClassroomFromCtx(r)
	if classroom == nil {
		app.notfoundResponse(w, r, fmt.Errorf("classroom not found"))
		return
	}

	app.jsonResponse(w, http.StatusOK, classroom)
}

// updateClassroomHandler
func (app *application) updateClassroomHandler(w http.ResponseWriter, r *http.Request) {
	classroom := getClassroomFromCtx(r)
	if classroom == nil {
		app.notfoundResponse(w, r, fmt.Errorf("classroom not found"))
		return
	}

	var payload UpdateClassroomPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	utils.ApplyPatch(classroom, payload)

	if err := app.store.Classrooms.Update(r.Context(), classroom); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notfoundResponse(w, r, err)
		default:
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}

	app.jsonResponse(w, http.StatusOK, classroom)
}

// deleteClassroomHandler
func (app *application) deleteClassroomHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "classroomID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Classrooms.Delete(r.Context(), id); err != nil {
		switch {
		case err == store.ErrNotFound:
			app.notfoundResponse(w, r, err)
		default:
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ------------------- Middleware -------------------

func (app *application) classroomsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "classroomID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		classroom, err := app.store.Classrooms.GetByID(r.Context(), id)
		if err != nil {
			switch {
			case err == store.ErrNotFound:
				app.notfoundResponse(w, r, err)
			default:
				app.internalServerErrorResponse(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), classroomCtx, classroom)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getClassroomFromCtx(r *http.Request) *store.Classroom {
	c, _ := r.Context().Value(classroomCtx).(*store.Classroom)
	return c
}
