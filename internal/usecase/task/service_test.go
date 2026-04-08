package task

//unit-тесты на recurrence
import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type stubRepository struct {
	created []taskdomain.Task
}

func (r *stubRepository) CreateMany(_ context.Context, tasks []taskdomain.Task) ([]taskdomain.Task, error) {
	r.created = append([]taskdomain.Task(nil), tasks...)
	for i := range r.created {
		r.created[i].ID = int64(i + 1)
	}
	return r.created, nil
}

func (r *stubRepository) GetByID(_ context.Context, _ int64) (*taskdomain.Task, error) {
	return nil, nil
}

func (r *stubRepository) Update(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	return task, nil
}

func (r *stubRepository) Delete(_ context.Context, _ int64) error {
	return nil
}

func (r *stubRepository) List(_ context.Context) ([]taskdomain.Task, error) {
	return nil, nil
}

func TestServiceCreateDailyRecurrence(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	}

	start := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)

	created, err := svc.Create(context.Background(), CreateInput{
		Title: "Call patients",
		Recurrence: &RecurrenceInput{
			Type:       RecurrenceTypeDaily,
			EveryNDays: 2,
			StartDate:  &start,
			EndDate:    &end,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(created) != 3 {
		t.Fatalf("len(created) = %d, want 3", len(created))
	}

	want := []string{"2026-04-10", "2026-04-12", "2026-04-14"}
	for i := range created {
		if got := created[i].ScheduleFor.Format(time.DateOnly); got != want[i] {
			t.Fatalf("created[%d].ScheduledFor = %s, want %s", i, got, want[i])
		}
	}
}

func TestServiceCreateSpecificDatesDeduplicates(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	}

	created, err := svc.Create(context.Background(), CreateInput{
		Title: "Check reports",
		Recurrence: &RecurrenceInput{
			Type: RecurrenceTypeSpecificDates,
			Dates: []time.Time{
				time.Date(2026, 4, 10, 15, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 12, 8, 0, 0, 0, time.UTC),
			},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(created) != 2 {
		t.Fatalf("len(created) = %d, want 2", len(created))
	}

	want := []string{"2026-04-10", "2026-04-12"}
	for i := range created {
		if got := created[i].ScheduleFor.Format(time.DateOnly); got != want[i] {
			t.Fatalf("created[%d].ScheduledFor = %s, want %s", i, got, want[i])
		}
	}
}

func TestServiceCreateRejectsScheduledForWithRecurrence(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)
	scheduledFor := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)

	_, err := svc.Create(context.Background(), CreateInput{
		Title:       "Call patients",
		ScheduleFor: &scheduledFor,
		Recurrence: &RecurrenceInput{
			Type:       RecurrenceTypeDaily,
			EveryNDays: 1,
			StartDate:  &start,
			EndDate:    &end,
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "scheduled_for cannot be used together with recurrence") {
		t.Fatalf("Create() error = %v, want scheduled_for conflict message", err)
	}
}

func TestServiceCreateRejectsInvalidStatus(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{
		Title:  "Call patients",
		Status: taskdomain.Status("bad_status"),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
	}
}

func TestRecurrenceNormalizedDatesValidation(t *testing.T) {
	tests := []struct {
		name       string
		recurrence RecurrenceInput
		wantErr    string
	}{
		{
			name: "daily with invalid interval",
			recurrence: RecurrenceInput{
				Type:       RecurrenceTypeDaily,
				EveryNDays: 0,
				StartDate:  datePtr("2026-04-01"),
				EndDate:    datePtr("2026-04-10"),
			},
			wantErr: "every_n_days must be positive",
		},
		{
			name: "monthly with invalid day",
			recurrence: RecurrenceInput{
				Type:       RecurrenceTypeMonthly,
				DayOfMonth: 31,
				StartDate:  datePtr("2026-04-01"),
				EndDate:    datePtr("2026-04-30"),
			},
			wantErr: "day_of_month must be between 1 and 30",
		},
		{
			name: "specific dates empty",
			recurrence: RecurrenceInput{
				Type: RecurrenceTypeSpecificDates,
			},
			wantErr: "dates are required for specific_dates recurrence",
		},
		{
			name: "start date after end date",
			recurrence: RecurrenceInput{
				Type:      RecurrenceTypeOddDays,
				StartDate: datePtr("2026-04-10"),
				EndDate:   datePtr("2026-04-01"),
			},
			wantErr: "start_date must be less than or equal to end_date",
		},
		{
			name: "range without matching dates",
			recurrence: RecurrenceInput{
				Type:       RecurrenceTypeMonthly,
				DayOfMonth: 30,
				StartDate:  datePtr("2026-02-01"),
				EndDate:    datePtr("2026-02-28"),
			},
			wantErr: "recurrence does not create any tasks in selected period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.recurrence.NormalizedDates()
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("NormalizedDates() error = %v, want ErrInvalidInput", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NormalizedDates() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func datePtr(value string) *time.Time {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		panic(err)
	}
	return &parsed
}
