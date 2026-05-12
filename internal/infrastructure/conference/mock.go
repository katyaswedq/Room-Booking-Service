package conference

import (
	"context"
	"fmt"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type MockService struct{}

func NewMockService() *MockService {
	return &MockService{}
}

func (s *MockService) CreateLink(ctx context.Context, slot domainslot.Slot) (string, error) {
	return fmt.Sprintf("https://conference.local/rooms/%s/slots/%s", slot.RoomID, slot.ID), nil
}

var _ domainbooking.ConferenceService = (*MockService)(nil)