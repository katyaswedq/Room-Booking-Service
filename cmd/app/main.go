package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	dbPool, err := database.NewPool(dbCtx, cfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer dbPool.Close()

	log.Println("postgres connected")

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

	conferenceService := conference.NewMockService()

	createRoomUseCase := roomusecase.NewCreateRoomUseCase(roomRepo)
	listRoomsUseCase := roomusecase.NewListRoomsUseCase(roomRepo)

	createScheduleUseCase := scheduleusecase.NewCreateUseCase(trManager, scheduleRepo, roomRepo, slotRepo)
	listSlotsUseCase := slotusecase.NewListUseCase(slotRepo, roomRepo)

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

	addr := ":" + cfg.AppPort

	server := &http.Server{
		Addr:              addr,
		Handler:           routerHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("starting server on %s", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := server.Close(); err != nil {
			log.Printf("server close failed: %v", err)
		}
	}

	log.Println("server stopped gracefully")
}