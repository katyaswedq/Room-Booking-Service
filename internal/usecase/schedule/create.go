package schedule

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
	domainschedule "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/schedule"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
	slotusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/slot"
)

const slotGenerationDays = 7

var (
	ErrInvalidDaysOfWeek = errors.New("invalid days of week")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrRoomNotFound      = errors.New("room not found")
	ErrScheduleExists    = errors.New("schedule already exists")
)

type CreateUseCase struct {
	trManager    trm.Manager
	scheduleRepo domainschedule.Repository
	roomRepo     domainroom.Repository
	slotRepo     domainslot.Repository
}

type CreateInput struct {
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type CreateOutput struct {
	ID         string
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

func NewCreateUseCase(trManager trm.Manager, scheduleRepo domainschedule.Repository, roomRepo domainroom.Repository, slotRepo domainslot.Repository) *CreateUseCase {
	return &CreateUseCase{
		trManager:    trManager,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		slotRepo:     slotRepo,
	}
}

func (uc *CreateUseCase) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	roomID := strings.TrimSpace(input.RoomID)
	startTime := strings.TrimSpace(input.StartTime)
	endTime := strings.TrimSpace(input.EndTime)

	if roomID == "" {
		return nil, ErrRoomNotFound
	}
	if _, err := uuid.Parse(roomID); err != nil {
		return nil, ErrRoomNotFound
	}

	if len(input.DaysOfWeek) == 0 {
		return nil, ErrInvalidDaysOfWeek
	}

	for _, day := range input.DaysOfWeek {
		if day < 1 || day > 7 {
			return nil, ErrInvalidDaysOfWeek
		}
	}

	startParsed, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, ErrInvalidTimeRange
	}

	endParsed, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, ErrInvalidTimeRange
	}

	if !startParsed.Before(endParsed) {
		return nil, ErrInvalidTimeRange
	}

	roomExists, err := uc.roomRepo.ExistsByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !roomExists {
		return nil, ErrRoomNotFound
	}

	schedule := &domainschedule.Schedule{
		ID:         uuid.NewString(),
		RoomID:     roomID,
		DaysOfWeek: input.DaysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	slots, err := slotusecase.GenerateForNextDays(roomID, input.DaysOfWeek, startParsed, endParsed, time.Now().UTC(), slotGenerationDays)
	if err != nil {
		if errors.Is(err, slotusecase.ErrInvalidTimeRange) {
			return nil, ErrInvalidTimeRange
		}
		return nil, err
	}

	err = uc.trManager.Do(ctx, func(ctx context.Context) error {
		scheduleExists, err := uc.scheduleRepo.ExistsByRoomID(ctx, roomID)
		if err != nil {
			return err
		}
		if scheduleExists {
			return ErrScheduleExists
		}

		if err := uc.scheduleRepo.Create(ctx, schedule); err != nil {
			return err
		}

		if err := uc.slotRepo.BulkCreate(ctx, slots); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreateOutput{
		ID:         schedule.ID,
		RoomID:     schedule.RoomID,
		DaysOfWeek: schedule.DaysOfWeek,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}, nil
}
