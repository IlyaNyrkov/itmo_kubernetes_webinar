package router

import (
	"app/middleware"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/time", middleware.GetTime).Methods("GET")

	return router
}
