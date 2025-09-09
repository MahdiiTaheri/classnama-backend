package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type markAttendancePayload struct {
	StudentID   int64   `json:"student_id" validate:"required"`
	TeacherID   *int64  `json:"teacher_id,omitempty"`
	ClassroomID *int64  `json:"classroom_id,omitempty"`
	Date        string  `json:"date" validate:"required,datetime=2006-01-02"`
	Status      string  `json:"status" validate:"required,oneof=present absent late excused"`
	Note        *string `json:"note,omitempty"`
}

type bulkAttendanceItem struct {
	StudentID int64  `json:"student_id" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=present absent late excused"`
}

type bulkAttendancePayload struct {
	ClassroomID int64                `json:"classroom_id" validate:"required"`
	Date        string               `json:"date" validate:"required,datetime=2006-01-02"`
	Statuses    []bulkAttendanceItem `json:"statuses" validate:"required,dive"`
}

// POST /api/attendance
// MarkAttendance godoc
//
//	@Summary	Mark a single attendance record (create or update)
//	@Tags		Attendance
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		markAttendancePayload	true	"Attendance payload"
//	@Success	201		{object}	store.AttendanceRecord
//	@Failure	400		{object}	error
//	@Failure	500		{object}	error
//	@Security	ApiKeyAuth
//	@Router		/attendance [post]
//	@ID			markAttendance
func (app *application) markAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	var payload markAttendancePayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// parse date
	dt, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid date format; expected YYYY-MM-DD"))
		return
	}

	rec := &store.AttendanceRecord{
		StudentID:   payload.StudentID,
		Date:        dt,
		Status:      payload.Status,
		TeacherID:   payload.TeacherID,
		ClassroomID: payload.ClassroomID,
		Note:        payload.Note,
	}

	if err := app.store.Attendance.Mark(r.Context(), rec); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, rec); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// POST /api/attendance/bulk
// BulkMarkAttendance godoc
//
//	@Summary	Bulk mark attendance for a classroom
//	@Tags		Attendance
//	@Accept		json
//	@Produce	json
//	@Param		payload	body	bulkAttendancePayload	true	"Bulk attendance payload"
//	@Success	204
//	@Failure	400	{object}	error
//	@Failure	500	{object}	error
//	@Security	ApiKeyAuth
//	@Router		/attendance/bulk [post]
//	@ID			bulkMarkAttendance
func (app *application) bulkMarkAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	var payload bulkAttendancePayload
	if err := readJSON(w, r, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dt, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid date format; expected YYYY-MM-DD"))
		return
	}

	statusMap := make(map[int64]string, len(payload.Statuses))
	for _, it := range payload.Statuses {
		statusMap[it.StudentID] = it.Status
	}

	if err := app.store.Attendance.BulkMark(r.Context(), payload.ClassroomID, dt, statusMap); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/attendance/students/{studentID}?from=&to=
// GetAttendanceByStudent godoc
//
//	@Summary	Get attendance records for a student
//	@Tags		Attendance
//	@Produce	json
//	@Param		studentID	path		int		true	"Student ID"
//	@Param		from		query		string	false	"From date YYYY-MM-DD"
//	@Param		to			query		string	false	"To date YYYY-MM-DD"
//	@Success	200			{array}		store.AttendanceRecord
//	@Failure	400			{object}	error
//	@Failure	404			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/attendance/students/{studentID} [get]
//	@ID			getAttendanceByStudent
func (app *application) getAttendanceByStudentHandler(w http.ResponseWriter, r *http.Request) {
	studentParam := chi.URLParam(r, "studentID")
	studentID, err := strconv.ParseInt(studentParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid student ID"))
		return
	}

	q := r.URL.Query()
	var from *time.Time
	if f := q.Get("from"); f != "" {
		t, err := time.Parse("2006-01-02", f)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid 'from' date"))
			return
		}
		from = &t
	}
	var to *time.Time
	if tstr := q.Get("to"); tstr != "" {
		t, err := time.Parse("2006-01-02", tstr)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid 'to' date"))
			return
		}
		to = &t
	}

	records, err := app.store.Attendance.GetByStudent(r.Context(), studentID, from, to)
	if err != nil {
		// treat no rows as not found? store returns empty slice for none; handle error
		if errors.Is(err, store.ErrNotFound) {
			app.notfoundResponse(w, r, err)
			return
		}
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if len(records) == 0 {
		// return empty array with 200 OR 404 based on your conventions; here return 200 empty list
		if err := app.jsonResponse(w, http.StatusOK, []store.AttendanceRecord{}); err != nil {
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, records); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}

// GET /api/attendance/classrooms/{classroomID}?date=YYYY-MM-DD
// GetAttendanceByClassroomDate godoc
//
//	@Summary	Get attendance for a classroom on a date
//	@Tags		Attendance
//	@Produce	json
//	@Param		classroomID	path		int		true	"Classroom ID"
//	@Param		date		query		string	true	"Date YYYY-MM-DD"
//	@Success	200			{array}		store.AttendanceRecord
//	@Failure	400			{object}	error
//	@Failure	500			{object}	error
//	@Security	ApiKeyAuth
//	@Router		/attendance/classrooms/{classroomID} [get]
//	@ID			getAttendanceByClassroomDate
func (app *application) getAttendanceByClassroomDateHandler(w http.ResponseWriter, r *http.Request) {
	classParam := chi.URLParam(r, "classroomID")
	classID, err := strconv.ParseInt(classParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid classroom ID"))
		return
	}

	q := r.URL.Query()
	dateStr := q.Get("date")
	if dateStr == "" {
		app.badRequestResponse(w, r, fmt.Errorf("missing date param (YYYY-MM-DD)"))
		return
	}
	dt, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid date param (YYYY-MM-DD)"))
		return
	}

	records, err := app.store.Attendance.GetByClassroomDate(r.Context(), classID, dt)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notfoundResponse(w, r, err)
			return
		}
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, records); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
}
