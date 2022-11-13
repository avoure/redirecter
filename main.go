package main

import (
	"net/http"
	"redirecter/pkg/db"
	"redirecter/pkg/handlers"

	"github.com/gorilla/mux"
)

func main() {
	DB := db.Init()
	h := handlers.NewHandler(DB)
	router := mux.NewRouter()
	router.HandleFunc("/links", h.GetAllLinks).Methods(http.MethodGet)
	router.HandleFunc("/links/{id}", h.GetLink).Methods(http.MethodGet)
	router.HandleFunc("/links/{id}", h.DeleteLink).Methods(http.MethodDelete)
	router.HandleFunc("/links", h.CreateLink).Methods(http.MethodPost)
	http.ListenAndServe(":8090", router)
}
