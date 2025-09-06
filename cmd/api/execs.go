package main

import (
	"errors"
	"net/http"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
)

// type execKey string

// const execCtx execKey = "exec"

type CreateExecPayload struct {
	FirstName string     `json:"first_name" validate:"required,max=72"`
	LastName  string     `json:"last_name" validate:"required,max=72"`
	Role      store.Role `json:"role"`
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
