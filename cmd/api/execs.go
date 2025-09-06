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

type execKey string

const execCtx execKey = "exec"

type CreateExecPayload struct {
	FirstName string     `json:"first_name" validate:"required,max=72"`
	LastName  string     `json:"last_name" validate:"required,max=72"`
	Email     string     `json:"email" validate:"required,email"`
	Role      store.Role `json:"role" validate:"required,oneof=admin manager"`
}

type UpdateExecPayload struct {
	FirstName *string     `json:"first_name,omitempty" validate:"omitempty,max=72"`
	LastName  *string     `json:"last_name,omitempty" validate:"omitempty,max=72"`
	Email     *string     `json:"email,omitempty" validate:"omitempty,email"`
	Role      *store.Role `json:"role,omitempty" validate:"omitempty,oneof=admin manager"`
}

func (app *application) createExecHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateExecPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	exec := &store.Exec{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Role:      payload.Role,
		Email:     payload.Email,
	}
	ctx := r.Context()

	if err := app.store.Execs.Create(ctx, exec); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, exec); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

}

func (app *application) getExecsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	execs, err := app.store.Execs.GetAll(ctx)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, execs); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

func (app *application) getExecHandler(w http.ResponseWriter, r *http.Request) {
	exec := getExecFromCtx(r)
	if exec == nil {
		app.notfoundResponse(w, r, fmt.Errorf("exec not found in context"))
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, exec); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

func (app *application) execsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "execID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()

		exec, err := app.store.Execs.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notfoundResponse(w, r, err)
			default:
				app.internalServerErrorResponse(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, execCtx, exec)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) updateExecHandler(w http.ResponseWriter, r *http.Request) {
	exec := getExecFromCtx(r)
	if exec == nil {
		app.notfoundResponse(w, r, fmt.Errorf("exec not found"))
		return
	}

	var payload UpdateExecPayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Apply non-nil fields using reflection
	utils.ApplyPatch(exec, payload)

	// Update in DB
	if err := app.store.Execs.Update(r.Context(), exec); err != nil {
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

	// Return updated exec
	if err := app.jsonResponse(w, http.StatusOK, exec); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

func getExecFromCtx(r *http.Request) *store.Exec {
	exec, _ := r.Context().Value(execCtx).(*store.Exec)
	return exec
}
