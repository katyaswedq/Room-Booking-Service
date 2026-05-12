package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/auth"
	bookinghandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/booking"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/middleware"
	roomhandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/room"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/router"
	schedulehandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/schedule"
	slothandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/slot"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/config"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/conference"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/database"
	"github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/repo"
	authusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/auth"
	bookingusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/booking"
	roomusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/room"
	scheduleusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/schedule"
	slotusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/slot"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestServer(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()

	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5433",
		DBName:     "room_booking_test",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBSSLMode:  "disable",
		JWTSecret:  "super-secret-key",
		AppPort:    "8080",
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	dbPool, err := database.NewPool(dbCtx, cfg)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	t.Cleanup(func() {
		dbPool.Close()
	})

	ctxGetter := trmpgx.DefaultCtxGetter
	trManager := manager.Must(trmpgx.NewDefaultFactory(dbPool))

	dummyLoginUseCase := authusecase.NewDummyLoginUseCase(cfg.JWTSecret)
	dummyLoginHandler := authhandler.NewHandler(dummyLoginUseCase)

	userRepo := repo.NewUserRepository(dbPool)
	registerUseCase := authusecase.NewRegisterUseCase(userRepo)
	registerHandler := authhandler.NewRegisterHandler(registerUseCase)
	loginUseCase := authusecase.NewLoginUseCase(userRepo, cfg.JWTSecret)
	loginHandler := authhandler.NewLoginHandler(loginUseCase)

	roomRepo := repo.NewRoomRepository(dbPool)
	scheduleRepo := repo.NewScheduleRepository(dbPool, ctxGetter)
	slotRepo := repo.NewSlotRepository(dbPool, ctxGetter)
	bookingRepo := repo.NewBookingRepository(dbPool, ctxGetter)

	// 👇 новое
	conferenceService := conference.NewMockService()

	createRoomUseCase := roomusecase.NewCreateRoomUseCase(roomRepo)
	listRoomsUseCase := roomusecase.NewListRoomsUseCase(roomRepo)

	createScheduleUseCase := scheduleusecase.NewCreateUseCase(trManager, scheduleRepo, roomRepo, slotRepo)
	listSlotsUseCase := slotusecase.NewListUseCase(slotRepo, roomRepo)

	// 👇 изменилось
	createBookingUseCase := bookingusecase.NewCreateUseCase(bookingRepo, slotRepo, conferenceService)

	cancelBookingUseCase := bookingusecase.NewCancelUseCase(bookingRepo)
	myBookingsUseCase := bookingusecase.NewMyUseCase(bookingRepo, slotRepo)
	listBookingsUseCase := bookingusecase.NewListUseCase(bookingRepo)

	createRoomHandler := roomhandler.NewCreateHandler(createRoomUseCase)
	listRoomsHandler := roomhandler.NewListHandler(listRoomsUseCase)

	createScheduleHandler := schedulehandler.NewCreateHandler(createScheduleUseCase)
	listSlotsHandler := slothandler.NewListHandler(listSlotsUseCase)

	createBookingHandler := bookinghandler.NewCreateHandler(createBookingUseCase)
	cancelBookingHandler := bookinghandler.NewCancelHandler(cancelBookingUseCase)
	myBookingsHandler := bookinghandler.NewMyHandler(myBookingsUseCase)
	listBookingsHandler := bookinghandler.NewListHandler(listBookingsUseCase)

	jwtMiddleware := middleware.JWT(cfg.JWTSecret)

	routerHandler := router.New(
		jwtMiddleware,
		registerHandler.ServeHTTP,
		loginHandler.ServeHTTP,
		dummyLoginHandler.DummyLogin,
		createRoomHandler.ServeHTTP,
		listRoomsHandler.ServeHTTP,
		createScheduleHandler.ServeHTTP,
		listSlotsHandler.ServeHTTP,
		createBookingHandler.ServeHTTP,
		cancelBookingHandler.ServeHTTP,
		myBookingsHandler.ServeHTTP,
		listBookingsHandler.ServeHTTP,
	)

	return httptest.NewServer(routerHandler), dbPool
}

func cleanDB(t *testing.T, dbPool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := dbPool.Exec(ctx, `
	TRUNCATE TABLE bookings, slots, schedules, rooms, users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func TestPing(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	resp, err := http.Get(server.URL + "/_info")
	if err != nil {
		t.Fatalf("failed to call _info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestBookingFlow_CreateRoomScheduleAndBooking(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	testDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")

	// 1. dummyLogin as admin
	loginReqBody := []byte(`{"role":"admin"}`)

	loginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader(loginReqBody),
	)
	if err != nil {
		t.Fatalf("dummyLogin admin request failed: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		t.Fatalf("expected 200 from dummyLogin, got %d, body: %s", loginResp.StatusCode, string(body))
	}

	var adminLoginResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(loginResp.Body).Decode(&adminLoginResp); err != nil {
		t.Fatalf("decode dummyLogin response: %v", err)
	}

	if adminLoginResp.Token == "" {
		t.Fatal("expected non-empty admin token")
	}

	// 2. create room as admin
	createRoomReqBody := []byte(`{"name":"Room A","description":"First room","capacity":6}`)

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader(createRoomReqBody),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminLoginResp.Token)

	client := &http.Client{}
	createRoomResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	if createRoomResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createRoomResp.Body)
		t.Fatalf("expected 201 from create room, got %d, body: %s", createRoomResp.StatusCode, string(body))
	}

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}

	if err := json.NewDecoder(createRoomResp.Body).Decode(&roomResp); err != nil {
		t.Fatalf("decode create room response: %v", err)
	}

	if roomResp.Room.ID == "" {
		t.Fatal("expected created room id")
	}

	// 3. create schedule as admin
	createScheduleReqBody := []byte(`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`)

	scheduleReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader(createScheduleReqBody),
	)
	if err != nil {
		t.Fatalf("create request for schedule: %v", err)
	}

	scheduleReq.Header.Set("Content-Type", "application/json")
	scheduleReq.Header.Set("Authorization", "Bearer "+adminLoginResp.Token)

	createScheduleResp, err := client.Do(scheduleReq)
	if err != nil {
		t.Fatalf("create schedule request failed: %v", err)
	}
	defer createScheduleResp.Body.Close()

	if createScheduleResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createScheduleResp.Body)
		t.Fatalf("expected 201 from create schedule, got %d, body: %s", createScheduleResp.StatusCode, string(body))
	}

	var scheduleResp struct {
		Schedule struct {
			ID string `json:"id"`
		} `json:"schedule"`
	}

	if err := json.NewDecoder(createScheduleResp.Body).Decode(&scheduleResp); err != nil {
		t.Fatalf("decode create schedule response: %v", err)
	}

	if scheduleResp.Schedule.ID == "" {
		t.Fatal("expected created schedule id")
	}

	// 4. list slots
	listSlotsReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list: %v", err)
	}

	listSlotsReq.Header.Set("Authorization", "Bearer "+adminLoginResp.Token)

	listSlotsResp, err := client.Do(listSlotsReq)
	if err != nil {
		t.Fatalf("list slots request failed: %v", err)
	}
	defer listSlotsResp.Body.Close()

	if listSlotsResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listSlotsResp.Body)
		t.Fatalf("expected 200 from list slots, got %d, body: %s", listSlotsResp.StatusCode, string(body))
	}

	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}

	if err := json.NewDecoder(listSlotsResp.Body).Decode(&slotsResp); err != nil {
		t.Fatalf("decode list slots response: %v", err)
	}

	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected non-empty slots list")
	}

	// 5. dummyLogin as user
	userLoginReqBody := []byte(`{"role":"user"}`)

	userLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader(userLoginReqBody),
	)
	if err != nil {
		t.Fatalf("dummyLogin user request failed: %v", err)
	}
	defer userLoginResp.Body.Close()

	if userLoginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userLoginResp.Body)
		t.Fatalf("expected 200 from user dummyLogin, got %d, body: %s", userLoginResp.StatusCode, string(body))
	}

	var userResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(userLoginResp.Body).Decode(&userResp); err != nil {
		t.Fatalf("decode user dummyLogin response: %v", err)
	}

	if userResp.Token == "" {
		t.Fatal("expected non-empty user token")
	}

	// 6. create booking as user
	bookedSlotID := slotsResp.Slots[0].ID
	createBookingReqBody := []byte(`{"slotId":"` + bookedSlotID + `"}`)

	createBookingReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/create",
		bytes.NewReader(createBookingReqBody),
	)
	if err != nil {
		t.Fatalf("create request for booking: %v", err)
	}

	createBookingReq.Header.Set("Content-Type", "application/json")
	createBookingReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	createBookingResp, err := client.Do(createBookingReq)
	if err != nil {
		t.Fatalf("create booking request failed: %v", err)
	}
	defer createBookingResp.Body.Close()

	if createBookingResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createBookingResp.Body)
		t.Fatalf("expected 201 from create booking, got %d, body: %s", createBookingResp.StatusCode, string(body))
	}

	var bookingResp struct {
		Booking struct {
			ID     string `json:"id"`
			SlotID string `json:"slotId"`
			UserID string `json:"userId"`
			Status string `json:"status"`
		} `json:"booking"`
	}

	if err := json.NewDecoder(createBookingResp.Body).Decode(&bookingResp); err != nil {
		t.Fatalf("decode create booking response: %v", err)
	}

	if bookingResp.Booking.ID == "" {
		t.Fatal("expected created booking id")
	}

	if bookingResp.Booking.SlotID != bookedSlotID {
		t.Fatalf("expected booked slot id %s, got %s", bookedSlotID, bookingResp.Booking.SlotID)
	}

	if bookingResp.Booking.Status != "active" {
		t.Fatalf("expected booking status active, got %s", bookingResp.Booking.Status)
	}

	// 7. list slots again and ensure booked slot disappeared
	listSlotsAfterBookingReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list after booking: %v", err)
	}

	listSlotsAfterBookingReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	listSlotsAfterBookingResp, err := client.Do(listSlotsAfterBookingReq)
	if err != nil {
		t.Fatalf("list slots after booking request failed: %v", err)
	}
	defer listSlotsAfterBookingResp.Body.Close()

	if listSlotsAfterBookingResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listSlotsAfterBookingResp.Body)
		t.Fatalf("expected 200 from list slots after booking, got %d, body: %s", listSlotsAfterBookingResp.StatusCode, string(body))
	}

	var slotsAfterBookingResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}

	if err := json.NewDecoder(listSlotsAfterBookingResp.Body).Decode(&slotsAfterBookingResp); err != nil {
		t.Fatalf("decode list slots after booking response: %v", err)
	}

	for _, slot := range slotsAfterBookingResp.Slots {
		if slot.ID == bookedSlotID {
			t.Fatalf("expected booked slot %s to disappear from available slots", bookedSlotID)
		}
	}
}

func TestBookingFlow_CancelBooking(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	testDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	client := &http.Client{}

	// 1. dummyLogin as admin
	adminLoginReqBody := []byte(`{"role":"admin"}`)

	adminLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader(adminLoginReqBody),
	)
	if err != nil {
		t.Fatalf("dummyLogin admin request failed: %v", err)
	}
	defer adminLoginResp.Body.Close()

	if adminLoginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(adminLoginResp.Body)
		t.Fatalf("expected 200 from admin dummyLogin, got %d, body: %s", adminLoginResp.StatusCode, string(body))
	}

	var adminResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(adminLoginResp.Body).Decode(&adminResp); err != nil {
		t.Fatalf("decode admin dummyLogin response: %v", err)
	}

	if adminResp.Token == "" {
		t.Fatal("expected non-empty admin token")
	}

	// 2. create room as admin
	createRoomReqBody := []byte(`{"name":"Room A","description":"First room","capacity":6}`)

	createRoomReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader(createRoomReqBody),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}

	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createRoomResp, err := client.Do(createRoomReq)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	if createRoomResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createRoomResp.Body)
		t.Fatalf("expected 201 from create room, got %d, body: %s", createRoomResp.StatusCode, string(body))
	}

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}

	if err := json.NewDecoder(createRoomResp.Body).Decode(&roomResp); err != nil {
		t.Fatalf("decode create room response: %v", err)
	}

	if roomResp.Room.ID == "" {
		t.Fatal("expected created room id")
	}

	// 3. create schedule as admin
	createScheduleReqBody := []byte(`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`)

	createScheduleReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader(createScheduleReqBody),
	)
	if err != nil {
		t.Fatalf("create request for schedule: %v", err)
	}

	createScheduleReq.Header.Set("Content-Type", "application/json")
	createScheduleReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createScheduleResp, err := client.Do(createScheduleReq)
	if err != nil {
		t.Fatalf("create schedule request failed: %v", err)
	}
	defer createScheduleResp.Body.Close()

	if createScheduleResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createScheduleResp.Body)
		t.Fatalf("expected 201 from create schedule, got %d, body: %s", createScheduleResp.StatusCode, string(body))
	}

	// 4. list slots
	listSlotsReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list: %v", err)
	}

	listSlotsReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	listSlotsResp, err := client.Do(listSlotsReq)
	if err != nil {
		t.Fatalf("list slots request failed: %v", err)
	}
	defer listSlotsResp.Body.Close()

	if listSlotsResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listSlotsResp.Body)
		t.Fatalf("expected 200 from list slots, got %d, body: %s", listSlotsResp.StatusCode, string(body))
	}

	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}

	if err := json.NewDecoder(listSlotsResp.Body).Decode(&slotsResp); err != nil {
		t.Fatalf("decode list slots response: %v", err)
	}

	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected non-empty slots list")
	}

	bookedSlotID := slotsResp.Slots[0].ID

	// 5. dummyLogin as user
	userLoginReqBody := []byte(`{"role":"user"}`)

	userLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader(userLoginReqBody),
	)
	if err != nil {
		t.Fatalf("dummyLogin user request failed: %v", err)
	}
	defer userLoginResp.Body.Close()

	if userLoginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userLoginResp.Body)
		t.Fatalf("expected 200 from user dummyLogin, got %d, body: %s", userLoginResp.StatusCode, string(body))
	}

	var userResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(userLoginResp.Body).Decode(&userResp); err != nil {
		t.Fatalf("decode user dummyLogin response: %v", err)
	}

	if userResp.Token == "" {
		t.Fatal("expected non-empty user token")
	}

	// 6. create booking as user
	createBookingReqBody := []byte(`{"slotId":"` + bookedSlotID + `"}`)

	createBookingReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/create",
		bytes.NewReader(createBookingReqBody),
	)
	if err != nil {
		t.Fatalf("create request for booking: %v", err)
	}

	createBookingReq.Header.Set("Content-Type", "application/json")
	createBookingReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	createBookingResp, err := client.Do(createBookingReq)
	if err != nil {
		t.Fatalf("create booking request failed: %v", err)
	}
	defer createBookingResp.Body.Close()

	if createBookingResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createBookingResp.Body)
		t.Fatalf("expected 201 from create booking, got %d, body: %s", createBookingResp.StatusCode, string(body))
	}

	var bookingResp struct {
		Booking struct {
			ID     string `json:"id"`
			SlotID string `json:"slotId"`
			Status string `json:"status"`
		} `json:"booking"`
	}

	if err := json.NewDecoder(createBookingResp.Body).Decode(&bookingResp); err != nil {
		t.Fatalf("decode create booking response: %v", err)
	}

	if bookingResp.Booking.ID == "" {
		t.Fatal("expected created booking id")
	}

	if bookingResp.Booking.SlotID != bookedSlotID {
		t.Fatalf("expected booked slot id %s, got %s", bookedSlotID, bookingResp.Booking.SlotID)
	}

	bookingID := bookingResp.Booking.ID

	// 7. cancel booking
	cancelReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/"+bookingID+"/cancel",
		nil,
	)
	if err != nil {
		t.Fatalf("create request for cancel booking: %v", err)
	}

	cancelReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	cancelResp, err := client.Do(cancelReq)
	if err != nil {
		t.Fatalf("cancel booking request failed: %v", err)
	}
	defer cancelResp.Body.Close()

	if cancelResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(cancelResp.Body)
		t.Fatalf("expected 200 from cancel booking, got %d, body: %s", cancelResp.StatusCode, string(body))
	}

	var cancelRespBody struct {
		Booking struct {
			ID     string `json:"id"`
			SlotID string `json:"slotId"`
			Status string `json:"status"`
		} `json:"booking"`
	}

	if err := json.NewDecoder(cancelResp.Body).Decode(&cancelRespBody); err != nil {
		t.Fatalf("decode cancel booking response: %v", err)
	}

	if cancelRespBody.Booking.ID != bookingID {
		t.Fatalf("expected cancelled booking id %s, got %s", bookingID, cancelRespBody.Booking.ID)
	}

	if cancelRespBody.Booking.Status != "cancelled" {
		t.Fatalf("expected booking status cancelled, got %s", cancelRespBody.Booking.Status)
	}

	// 8. repeat cancel to verify idempotency
	repeatCancelReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/"+bookingID+"/cancel",
		nil,
	)
	if err != nil {
		t.Fatalf("create request for repeated cancel: %v", err)
	}

	repeatCancelReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	repeatCancelResp, err := client.Do(repeatCancelReq)
	if err != nil {
		t.Fatalf("repeat cancel booking request failed: %v", err)
	}
	defer repeatCancelResp.Body.Close()

	if repeatCancelResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(repeatCancelResp.Body)
		t.Fatalf("expected 200 from repeated cancel, got %d, body: %s", repeatCancelResp.StatusCode, string(body))
	}

	var repeatCancelRespBody struct {
		Booking struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"booking"`
	}

	if err := json.NewDecoder(repeatCancelResp.Body).Decode(&repeatCancelRespBody); err != nil {
		t.Fatalf("decode repeated cancel response: %v", err)
	}

	if repeatCancelRespBody.Booking.ID != bookingID {
		t.Fatalf("expected repeated cancelled booking id %s, got %s", bookingID, repeatCancelRespBody.Booking.ID)
	}

	if repeatCancelRespBody.Booking.Status != "cancelled" {
		t.Fatalf("expected repeated cancel status cancelled, got %s", repeatCancelRespBody.Booking.Status)
	}

	// 9. list slots again and ensure slot is available again
	listSlotsAfterCancelReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list after cancel: %v", err)
	}

	listSlotsAfterCancelReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	listSlotsAfterCancelResp, err := client.Do(listSlotsAfterCancelReq)
	if err != nil {
		t.Fatalf("list slots after cancel request failed: %v", err)
	}
	defer listSlotsAfterCancelResp.Body.Close()

	if listSlotsAfterCancelResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listSlotsAfterCancelResp.Body)
		t.Fatalf("expected 200 from list slots after cancel, got %d, body: %s", listSlotsAfterCancelResp.StatusCode, string(body))
	}

	var slotsAfterCancelResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}

	if err := json.NewDecoder(listSlotsAfterCancelResp.Body).Decode(&slotsAfterCancelResp); err != nil {
		t.Fatalf("decode list slots after cancel response: %v", err)
	}

	found := false
	for _, slot := range slotsAfterCancelResp.Slots {
		if slot.ID == bookedSlotID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected cancelled slot %s to appear in available slots again", bookedSlotID)
	}
}

func TestBookingFlow_CreateBooking_AsAdmin_ReturnsForbidden(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	testDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	client := &http.Client{}

	// admin login
	adminLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader([]byte(`{"role":"admin"}`)),
	)
	if err != nil {
		t.Fatalf("dummyLogin admin request failed: %v", err)
	}
	defer adminLoginResp.Body.Close()

	var adminResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(adminLoginResp.Body).Decode(&adminResp); err != nil {
		t.Fatalf("decode admin login response: %v", err)
	}

	// create room
	createRoomReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader([]byte(`{"name":"Room A","description":"First room","capacity":6}`)),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}
	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createRoomResp, err := client.Do(createRoomReq)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	if err := json.NewDecoder(createRoomResp.Body).Decode(&roomResp); err != nil {
		t.Fatalf("decode create room response: %v", err)
	}

	// create schedule
	createScheduleReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader([]byte(`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`)),
	)
	if err != nil {
		t.Fatalf("create request for schedule: %v", err)
	}
	createScheduleReq.Header.Set("Content-Type", "application/json")
	createScheduleReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createScheduleResp, err := client.Do(createScheduleReq)
	if err != nil {
		t.Fatalf("create schedule request failed: %v", err)
	}
	defer createScheduleResp.Body.Close()

	// list slots
	listSlotsReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list: %v", err)
	}
	listSlotsReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	listSlotsResp, err := client.Do(listSlotsReq)
	if err != nil {
		t.Fatalf("list slots request failed: %v", err)
	}
	defer listSlotsResp.Body.Close()

	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	if err := json.NewDecoder(listSlotsResp.Body).Decode(&slotsResp); err != nil {
		t.Fatalf("decode list slots response: %v", err)
	}
	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected non-empty slots list")
	}

	// try to create booking as admin -> must be 403
	createBookingReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/create",
		bytes.NewReader([]byte(`{"slotId":"`+slotsResp.Slots[0].ID+`"}`)),
	)
	if err != nil {
		t.Fatalf("create request for booking: %v", err)
	}
	createBookingReq.Header.Set("Content-Type", "application/json")
	createBookingReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createBookingResp, err := client.Do(createBookingReq)
	if err != nil {
		t.Fatalf("create booking request failed: %v", err)
	}
	defer createBookingResp.Body.Close()

	if createBookingResp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(createBookingResp.Body)
		t.Fatalf("expected 403 from create booking as admin, got %d, body: %s", createBookingResp.StatusCode, string(body))
	}
}

func TestScheduleFlow_CreateSchedule_Twice_ReturnsConflict(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	client := &http.Client{}

	adminLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader([]byte(`{"role":"admin"}`)),
	)
	if err != nil {
		t.Fatalf("dummyLogin admin request failed: %v", err)
	}
	defer adminLoginResp.Body.Close()

	var adminResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(adminLoginResp.Body).Decode(&adminResp); err != nil {
		t.Fatalf("decode admin login response: %v", err)
	}

	createRoomReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader([]byte(`{"name":"Room A","description":"First room","capacity":6}`)),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}
	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createRoomResp, err := client.Do(createRoomReq)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	if err := json.NewDecoder(createRoomResp.Body).Decode(&roomResp); err != nil {
		t.Fatalf("decode create room response: %v", err)
	}

	body := []byte(`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`)

	firstReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create first schedule request: %v", err)
	}
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	firstResp, err := client.Do(firstReq)
	if err != nil {
		t.Fatalf("first create schedule request failed: %v", err)
	}
	defer firstResp.Body.Close()

	if firstResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(firstResp.Body)
		t.Fatalf("expected 201 from first schedule creation, got %d, body: %s", firstResp.StatusCode, string(respBody))
	}

	secondReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create second schedule request: %v", err)
	}
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	secondResp, err := client.Do(secondReq)
	if err != nil {
		t.Fatalf("second create schedule request failed: %v", err)
	}
	defer secondResp.Body.Close()

	if secondResp.StatusCode != http.StatusConflict {
		respBody, _ := io.ReadAll(secondResp.Body)
		t.Fatalf("expected 409 from second schedule creation, got %d, body: %s", secondResp.StatusCode, string(respBody))
	}
}

func TestBookingFlow_CreateBooking_ForAlreadyBookedSlot_ReturnsConflict(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	testDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	client := &http.Client{}

	// admin login
	adminLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader([]byte(`{"role":"admin"}`)),
	)
	if err != nil {
		t.Fatalf("dummyLogin admin request failed: %v", err)
	}
	defer adminLoginResp.Body.Close()

	var adminResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(adminLoginResp.Body).Decode(&adminResp); err != nil {
		t.Fatalf("decode admin login response: %v", err)
	}

	// create room
	createRoomReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader([]byte(`{"name":"Room A","description":"First room","capacity":6}`)),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}
	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createRoomResp, err := client.Do(createRoomReq)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	if err := json.NewDecoder(createRoomResp.Body).Decode(&roomResp); err != nil {
		t.Fatalf("decode create room response: %v", err)
	}

	// create schedule
	createScheduleReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/"+roomResp.Room.ID+"/schedule/create",
		bytes.NewReader([]byte(`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`)),
	)
	if err != nil {
		t.Fatalf("create request for schedule: %v", err)
	}
	createScheduleReq.Header.Set("Content-Type", "application/json")
	createScheduleReq.Header.Set("Authorization", "Bearer "+adminResp.Token)

	createScheduleResp, err := client.Do(createScheduleReq)
	if err != nil {
		t.Fatalf("create schedule request failed: %v", err)
	}
	defer createScheduleResp.Body.Close()

	// user login
	userLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader([]byte(`{"role":"user"}`)),
	)
	if err != nil {
		t.Fatalf("dummyLogin user request failed: %v", err)
	}
	defer userLoginResp.Body.Close()

	var userResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(userLoginResp.Body).Decode(&userResp); err != nil {
		t.Fatalf("decode user login response: %v", err)
	}

	// list slots
	listSlotsReq, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/rooms/"+roomResp.Room.ID+"/slots/list?date="+testDate,
		nil,
	)
	if err != nil {
		t.Fatalf("create request for slots list: %v", err)
	}
	listSlotsReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	listSlotsResp, err := client.Do(listSlotsReq)
	if err != nil {
		t.Fatalf("list slots request failed: %v", err)
	}
	defer listSlotsResp.Body.Close()

	var slotsResp struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	if err := json.NewDecoder(listSlotsResp.Body).Decode(&slotsResp); err != nil {
		t.Fatalf("decode list slots response: %v", err)
	}
	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected non-empty slots list")
	}

	slotID := slotsResp.Slots[0].ID
	body := []byte(`{"slotId":"` + slotID + `"}`)

	// first booking -> 201
	firstBookingReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/create",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create first booking request: %v", err)
	}
	firstBookingReq.Header.Set("Content-Type", "application/json")
	firstBookingReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	firstBookingResp, err := client.Do(firstBookingReq)
	if err != nil {
		t.Fatalf("first booking request failed: %v", err)
	}
	defer firstBookingResp.Body.Close()

	if firstBookingResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(firstBookingResp.Body)
		t.Fatalf("expected 201 from first booking, got %d, body: %s", firstBookingResp.StatusCode, string(respBody))
	}

	// second booking same slot -> 409
	secondBookingReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/bookings/create",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create second booking request: %v", err)
	}
	secondBookingReq.Header.Set("Content-Type", "application/json")
	secondBookingReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	secondBookingResp, err := client.Do(secondBookingReq)
	if err != nil {
		t.Fatalf("second booking request failed: %v", err)
	}
	defer secondBookingResp.Body.Close()

	if secondBookingResp.StatusCode != http.StatusConflict {
		respBody, _ := io.ReadAll(secondBookingResp.Body)
		t.Fatalf("expected 409 from second booking, got %d, body: %s", secondBookingResp.StatusCode, string(respBody))
	}
}

func TestRoomFlow_CreateRoom_AsUser_ReturnsForbidden(t *testing.T) {
	server, dbPool := newTestServer(t)
	defer server.Close()

	cleanDB(t, dbPool)

	client := &http.Client{}

	userLoginResp, err := http.Post(
		server.URL+"/dummyLogin",
		"application/json",
		bytes.NewReader([]byte(`{"role":"user"}`)),
	)
	if err != nil {
		t.Fatalf("dummyLogin user request failed: %v", err)
	}
	defer userLoginResp.Body.Close()

	var userResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(userLoginResp.Body).Decode(&userResp); err != nil {
		t.Fatalf("decode user login response: %v", err)
	}

	createRoomReq, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/rooms/create",
		bytes.NewReader([]byte(`{"name":"Room A","description":"First room","capacity":6}`)),
	)
	if err != nil {
		t.Fatalf("create request for room: %v", err)
	}
	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+userResp.Token)

	createRoomResp, err := client.Do(createRoomReq)
	if err != nil {
		t.Fatalf("create room request failed: %v", err)
	}
	defer createRoomResp.Body.Close()

	if createRoomResp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(createRoomResp.Body)
		t.Fatalf("expected 403 from create room as user, got %d, body: %s", createRoomResp.StatusCode, string(body))
	}
}
