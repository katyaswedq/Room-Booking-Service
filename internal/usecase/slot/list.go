package slot

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

var (
	ErrInvalidDate  = errors.New("invalid date")
	ErrRoomNotFound = errors.New("room not found")
)

type ListUseCase struct {
	slotRepo domainslot.Repository
	roomRepo domainroom.Repository
}

type ListInput struct {
	RoomID string
	Date   string
}

type ListOutput struct {
	Slots []SlotOutput
}

type SlotOutput struct {
	ID     string
	RoomID string
	Start  time.Time
	End    time.Time
}

func NewListUseCase(slotRepo domainslot.Repository, roomRepo domainroom.Repository) *ListUseCase {
	return &ListUseCase{
		slotRepo: slotRepo,
		roomRepo: roomRepo,
	}
}

func (uc *ListUseCase) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	roomID := strings.TrimSpace(input.RoomID)
	date := strings.TrimSpace(input.Date)

	if roomID == "" {
		return nil, ErrRoomNotFound
	}

	if _, err := uuid.Parse(roomID); err != nil {
		return nil, ErrRoomNotFound
	}

	if date == "" {
		return nil, ErrInvalidDate
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	roomExists, err := uc.roomRepo.ExistsByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !roomExists {
		return nil, ErrRoomNotFound
	}

	slots, err := uc.slotRepo.ListAvailableByRoomAndDate(ctx, roomID, parsedDate)
	if err != nil {
		return nil, err
	}

	out := make([]SlotOutput, 0, len(slots))
	for _, slot := range slots {
		out = append(out, SlotOutput{
			ID:     slot.ID,
			RoomID: slot.RoomID,
			Start:  slot.Start,
			End:    slot.End,
		})
	}

	return &ListOutput{
		Slots: out,
	}, nil
}