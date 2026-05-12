package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/middleware"
	swaggerhandler "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/swagger"
)

// type Server struct {
// 	loginHandler http.HandlerFunc
// 	registerHandler http.HandlerFunc
// }

func New(
	jwtMiddleware func(http.Handler) http.Handler,
	registerHandler http.HandlerFunc,
	loginHandler http.HandlerFunc,
	dummyLoginHandler http.HandlerFunc,
	createRoomHandler http.HandlerFunc,
	listRoomsHandler http.HandlerFunc,
	createScheduleHandler http.HandlerFunc,
	listSlotsHandler http.HandlerFunc,
	createBookingHandler http.HandlerFunc,
	cancelBookingHandler http.HandlerFunc,
	myBookingsHandler http.HandlerFunc,
	listBookingsHandler http.HandlerFunc) http.Handler {
		
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/_info", info)
	r.Get("/swagger", swaggerhandler.UI())
	r.Get("/swagger/doc.yaml", swaggerhandler.Doc())

	r.Post("/register", registerHandler)
	r.Post("/login", loginHandler)
	r.Post("/dummyLogin", dummyLoginHandler)

	r.With(jwtMiddleware).Get("/rooms/list", listRoomsHandler)
	r.With(jwtMiddleware, middleware.RequireAdmin).Post("/rooms/create", createRoomHandler)
	r.With(jwtMiddleware, middleware.RequireAdmin).Post("/rooms/{roomId}/schedule/create", createScheduleHandler)
	r.With(jwtMiddleware).Get("/rooms/{roomId}/slots/list", listSlotsHandler)

	r.With(jwtMiddleware, middleware.RequireUser).Post("/bookings/create", createBookingHandler)
	r.With(jwtMiddleware, middleware.RequireUser).Post("/bookings/{bookingId}/cancel", cancelBookingHandler)
	r.With(jwtMiddleware, middleware.RequireUser).Get("/bookings/my", myBookingsHandler)
	r.With(jwtMiddleware, middleware.RequireAdmin).Get("/bookings/list", listBookingsHandler)

	return r
}

func info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}