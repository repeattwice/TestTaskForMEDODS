package handlers

//Добавлены HTTP-тесты для create с recurrence и ошибок
import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"github.com/gorilla/mux"
)

type stubUsecase struct {
	createFunc func(ctx context.Context, input taskusecase.CreateInput) ([]taskdomain.Task, error)
}

func (s *stubUsecase) Create(ctx context.Context, input taskusecase.CreateInput) ([]taskdomain.Task, error) {
	if s.createFunc != nil {
		return s.createFunc(ctx, input)
	}
	return nil, nil
}

func (s *stubUsecase) GetByID(context.Context, int64) (*taskdomain.Task, error) { return nil, nil }
func (s *stubUsecase) Update(context.Context, int64, taskusecase.UpdateInput) (*taskdomain.Task, error) {
	return nil, nil
}
func (s *stubUsecase) Delete(context.Context, int64) error             { return nil }
func (s *stubUsecase) List(context.Context) ([]taskdomain.Task, error) { return nil, nil }

func TestTaskHandlerCreateWithRecurrence(t *testing.T) {
	handler := NewTaskHandler(&stubUsecase{
		createFunc: func(_ context.Context, input taskusecase.CreateInput) ([]taskdomain.Task, error) {
			if input.Recurrence == nil {
				t.Fatalf("Create input recurrence = nil")
			}
			if input.Recurrence.Type != taskusecase.RecurrenceTypeDaily {
				t.Fatalf("Create input recurrence type = %q, want %q", input.Recurrence.Type, taskusecase.RecurrenceTypeDaily)
			}
			if input.Recurrence.EveryNDays != 2 {
				t.Fatalf("Create input recurrence every_n_days = %d, want 2", input.Recurrence.EveryNDays)
			}
			if got := input.Recurrence.StartDate.Format(time.DateOnly); got != "2026-04-01" {
				t.Fatalf("Create input recurrence start_date = %s, want 2026-04-01", got)
			}

			return []taskdomain.Task{
				{ID: 1, Title: input.Title, Description: input.Description, Status: taskdomain.StatusNew, ScheduleFor: mustDate(t, "2026-04-01"), CreatedAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)},
				{ID: 2, Title: input.Title, Description: input.Description, Status: taskdomain.StatusNew, ScheduleFor: mustDate(t, "2026-04-03"), CreatedAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)},
			}, nil
		},
	})

	body := `{
		"title":"Blood pressure check",
		"description":"Ward A",
		"recurrence":{
			"type":"daily",
			"every_n_days":2,
			"start_date":"2026-04-01",
			"end_date":"2026-04-03"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var response createTasksResponse
	if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.Tasks) != 2 {
		t.Fatalf("len(response.Tasks) = %d, want 2", len(response.Tasks))
	}
	if response.Tasks[0].ScheduledFor != "2026-04-01" || response.Tasks[1].ScheduledFor != "2026-04-03" {
		t.Fatalf("unexpected scheduled_for values: %+v", response.Tasks)
	}
}

func TestTaskHandlerCreateRejectsUnknownFields(t *testing.T) {
	handler := NewTaskHandler(&stubUsecase{})
	body := `{"title":"Blood pressure check","unknown":"field"}`

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "unknown field") {
		t.Fatalf("body = %s, want unknown field error", w.Body.String())
	}
}

func TestTaskHandlerCreateRejectsInvalidRecurrenceDate(t *testing.T) {
	handler := NewTaskHandler(&stubUsecase{})
	body := `{
		"title":"Blood pressure check",
		"recurrence":{
			"type":"specific_dates",
			"dates":["2026/04/01"]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "invalid recurrence date, expected YYYY-MM-DD") {
		t.Fatalf("body = %s, want invalid recurrence date error", w.Body.String())
	}
}

func TestTaskHandlerGetByIDRejectsNonPositiveID(t *testing.T) {
	handler := NewTaskHandler(&stubUsecase{})
	req := httptest.NewRequest(http.MethodGet, "/tasks/0", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "0"})
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "invalid task id") {
		t.Fatalf("body = %s, want invalid task id error", w.Body.String())
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}
