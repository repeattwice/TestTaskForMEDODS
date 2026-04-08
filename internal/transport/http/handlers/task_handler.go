package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/usecase/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskCreateDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toCreateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := createTasksResponse{Tasks: make([]taskDTO, 0, len(created))}
	for i := range created {
		response.Tasks = append(response.Tasks, newTaskDTO(&created[i]))
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskUpdateDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toUpdateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (dto taskCreateDTO) toCreateInput() (taskusecase.CreateInput, error) {
	scheduledFor, err := parseOptionalDate(dto.ScheduledFor)
	if err != nil {
		return taskusecase.CreateInput{}, err
	}

	recurrence, err := dto.Recurrence.toUsecaseInput()
	if err != nil {
		return taskusecase.CreateInput{}, err
	}

	return task.CreateInput{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
		ScheduleFor: scheduledFor,
		Recurrence:  recurrence,
	}, nil
}

func (dto taskUpdateDTO) toUpdateInput() (taskusecase.UpdateInput, error) {
	scheduledFor, err := parseOptionalDate(dto.ScheduledFor)
	if err != nil {
		return taskusecase.UpdateInput{}, err
	}

	return taskusecase.UpdateInput{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
		ScheduleFor: scheduledFor,
	}, nil
}

func (dto *recurrenceDTO) toUsecaseInput() (*taskusecase.RecurrenceInput, error) {
	if dto == nil {
		return nil, nil
	}

	startDate, err := parseOptionalDate(dto.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseOptionalDate(dto.EndDate)
	if err != nil {
		return nil, err
	}

	dates := make([]time.Time, 0, len(dto.Dates))
	for _, rawDate := range dto.Dates {
		date, err := time.Parse(time.DateOnly, rawDate)
		if err != nil {
			return nil, errors.New("invalid recurrence date, expected YYYY-MM-DD")
		}
		dates = append(dates, date)
	}

	return &taskusecase.RecurrenceInput{
		Type:       dto.Type,
		EveryNDays: dto.EveryNDays,
		DayOfMonth: dto.DayOfMonth,
		StartDate:  startDate,
		EndDate:    endDate,
		Dates:      dates,
	}, nil
}

func parseOptionalDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, errors.New("invalid date, expected YYYY-MM-DD")
	}
	return &parsed, nil
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}
	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
