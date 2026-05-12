package slot

import (
	"testing"
	"time"
)

func TestGenerateForNextDays_ForSingleDay_ReturnsThirtyMinuteSlots(t *testing.T) {
	roomID := "room-1"

	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC) // Wednesday
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC)

	slots, err := GenerateForNextDays(
		roomID,
		[]int{3}, // Wednesday
		startTime,
		endTime,
		now,
		1,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(slots) != 4 {
		t.Fatalf("expected 4 slots, got %d", len(slots))
	}

	expectedStarts := []time.Time{
		time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 9, 30, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 10, 30, 0, 0, time.UTC),
	}

	expectedEnds := []time.Time{
		time.Date(2026, 3, 25, 9, 30, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 10, 30, 0, 0, time.UTC),
		time.Date(2026, 3, 25, 11, 0, 0, 0, time.UTC),
	}

	for i, slot := range slots {
		if slot.ID == "" {
			t.Fatalf("expected slot %d to have non-empty id", i)
		}

		if slot.RoomID != roomID {
			t.Fatalf("expected slot %d roomID %q, got %q", i, roomID, slot.RoomID)
		}

		if !slot.Start.Equal(expectedStarts[i]) {
			t.Fatalf("expected slot %d start %v, got %v", i, expectedStarts[i], slot.Start)
		}

		if !slot.End.Equal(expectedEnds[i]) {
			t.Fatalf("expected slot %d end %v, got %v", i, expectedEnds[i], slot.End)
		}
	}
}

func TestGenerateForNextDays_DayNotAllowed_ReturnsNoSlots(t *testing.T) {
	roomID := "room-1"

	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC) 
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC)

	slots, err := GenerateForNextDays(
		roomID,
		[]int{1}, 
		startTime,
		endTime,
		now,
		1,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(slots) != 0 {
		t.Fatalf("expected 0 slots, got %d", len(slots))
	}
}

func TestGenerateForNextDays_MultipleDays_GeneratesOnlyAllowedDays(t *testing.T) {
	roomID := "room-1"

	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC) // Wednesday
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)

	slots, err := GenerateForNextDays(
		roomID,
		[]int{3, 5}, // Wednesday and Friday
		startTime,
		endTime,
		now,
		3, // Wed, Thu, Fri
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 2 allowed days * 2 slots per day (09:00-09:30, 09:30-10:00)
	if len(slots) != 4 {
		t.Fatalf("expected 4 slots, got %d", len(slots))
	}

	expectedStarts := []time.Time{
		time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC),  // Wed
		time.Date(2026, 3, 25, 9, 30, 0, 0, time.UTC), // Wed
		time.Date(2026, 3, 27, 9, 0, 0, 0, time.UTC),  // Fri
		time.Date(2026, 3, 27, 9, 30, 0, 0, time.UTC), // Fri
	}

	for i, slot := range slots {
		if slot.RoomID != roomID {
			t.Fatalf("expected slot %d roomID %q, got %q", i, roomID, slot.RoomID)
		}

		if !slot.Start.Equal(expectedStarts[i]) {
			t.Fatalf("expected slot %d start %v, got %v", i, expectedStarts[i], slot.Start)
		}
	}
}

func TestGenerateForNextDays_InvalidTimeRange_ReturnsError(t *testing.T) {
	roomID := "room-1"

	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)

	slots, err := GenerateForNextDays(
		roomID,
		[]int{3},
		startTime,
		endTime,
		now,
		1,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}

	if slots != nil {
		t.Fatalf("expected nil slots, got %v", slots)
	}
}