package schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
	domainschedule "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/schedule"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type mockTrManager struct {
	doFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTrManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.doFn != nil {
		return m.doFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockTrManager) DoWithSettings(
	ctx context.Context,
	_ trm.Settings,
	fn func(ctx context.Context) error,
) error {
	if m.doFn != nil {
		return m.doFn(ctx, fn)
	}
	return fn(ctx)
}

type mockScheduleRepository struct {
	existsByRoomIDFn func(ctx context.Context, roomID string) (bool, error)
	createFn         func(ctx context.Context, s *domainschedule.Schedule) error
}

func (m *mockScheduleRepository) Create(ctx context.Context, s *domainschedule.Schedule) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}

func (m *mockScheduleRepository) ExistsByRoomID(ctx context.Context, roomID string) (bool, error) {
	if m.existsByRoomIDFn != nil {
		return m.existsByRoomIDFn(ctx, roomID)
	}
	return false, nil
}

type mockRoomRepository struct {
	existsByIDFn func(ctx context.Context, id string) (bool, error)
}

func (m *mockRoomRepository) Create(ctx context.Context, room *domainroom.Room) error {
	return nil
}

func (m *mockRoomRepository) List(ctx context.Context) ([]domainroom.Room, error) {
	return nil, nil
}

func (m *mockRoomRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if m.existsByIDFn != nil {
		return m.existsByIDFn(ctx, id)
	}
	return false, nil
}

type mockSlotRepository struct {
	bulkCreateFn func(ctx context.Context, slots []domainslot.Slot) error
}

func (m *mockSlotRepository) BulkCreate(ctx context.Context, slots []domainslot.Slot) error {
	if m.bulkCreateFn != nil {
		return m.bulkCreateFn(ctx, slots)
	}
	return nil
}

func (m *mockSlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *mockSlotRepository) GetByID(ctx context.Context, id string) (*domainslot.Slot, error) {
	return nil, nil
}

func TestCreateSchedule_InvalidRoomID_ReturnsRoomNotFound(t *testing.T) {
	uc := NewCreateUseCase(&mockTrManager{}, &mockScheduleRepository{}, &mockRoomRepository{}, &mockSlotRepository{})

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     "bad-room-id",
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateSchedule_InvalidDaysOfWeek_ReturnsError(t *testing.T) {
	uc := NewCreateUseCase(&mockTrManager{}, &mockScheduleRepository{}, &mockRoomRepository{}, &mockSlotRepository{})

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     "11111111-1111-1111-1111-111111111111",
		DaysOfWeek: []int{0, 8},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidDaysOfWeek) {
		t.Fatalf("expected ErrInvalidDaysOfWeek, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateSchedule_InvalidTimeRange_ReturnsError(t *testing.T) {
	uc := NewCreateUseCase(&mockTrManager{}, &mockScheduleRepository{}, &mockRoomRepository{}, &mockSlotRepository{})

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     "11111111-1111-1111-1111-111111111111",
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "18:00",
		EndTime:    "09:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateSchedule_RoomNotFound_ReturnsError(t *testing.T) {
	roomRepo := &mockRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return false, nil
		},
	}

	uc := NewCreateUseCase(&mockTrManager{}, &mockScheduleRepository{}, roomRepo, &mockSlotRepository{})

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     "11111111-1111-1111-1111-111111111111",
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateSchedule_ScheduleExists_ReturnsError(t *testing.T) {
	roomRepo := &mockRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}

	scheduleRepo := &mockScheduleRepository{
		existsByRoomIDFn: func(ctx context.Context, roomID string) (bool, error) {
			return true, nil
		},
	}

	trManager := &mockTrManager{}

	uc := NewCreateUseCase(trManager, scheduleRepo, roomRepo, &mockSlotRepository{})

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     "11111111-1111-1111-1111-111111111111",
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrScheduleExists) {
		t.Fatalf("expected ErrScheduleExists, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateSchedule_Success(t *testing.T) {
	roomID := "11111111-1111-1111-1111-111111111111"

	roomRepo := &mockRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}

	var createdSchedule *domainschedule.Schedule
	scheduleRepo := &mockScheduleRepository{
		existsByRoomIDFn: func(ctx context.Context, roomID string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, s *domainschedule.Schedule) error {
			createdSchedule = s
			return nil
		},
	}

	bulkCreateCalled := false
	var createdSlots []domainslot.Slot
	slotRepo := &mockSlotRepository{
		bulkCreateFn: func(ctx context.Context, slots []domainslot.Slot) error {
			bulkCreateCalled = true
			createdSlots = slots
			return nil
		},
	}

	trManager := &mockTrManager{
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}

	uc := NewCreateUseCase(trManager, scheduleRepo, roomRepo, slotRepo)

	out, err := uc.Create(context.Background(), CreateInput{
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if createdSchedule == nil {
		t.Fatal("expected schedule to be created")
	}

	if createdSchedule.ID == "" {
		t.Fatal("expected created schedule ID to be set")
	}

	if createdSchedule.RoomID != roomID {
		t.Fatalf("expected roomID %s, got %s", roomID, createdSchedule.RoomID)
	}

	if !bulkCreateCalled {
		t.Fatal("expected BulkCreate to be called")
	}

	if len(createdSlots) == 0 {
		t.Fatal("expected generated slots to be created")
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.ID == "" {
		t.Fatal("expected output ID to be set")
	}

	if out.RoomID != roomID {
		t.Fatalf("expected output roomID %s, got %s", roomID, out.RoomID)
	}

	if out.StartTime != "09:00" {
		t.Fatalf("expected startTime 09:00, got %s", out.StartTime)
	}

	if out.EndTime != "18:00" {
		t.Fatalf("expected endTime 18:00, got %s", out.EndTime)
	}
}

var _ trm.Manager = (*mockTrManager)(nil)