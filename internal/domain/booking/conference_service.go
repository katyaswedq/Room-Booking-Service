package booking

import (
	"context"

	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type ConferenceService interface {
	CreateLink(ctx context.Context, slot domainslot.Slot) (string, error)
}