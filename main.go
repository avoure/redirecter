package main

import (
	"log"
	"net/http"
	"redirecter/pkg/db"
	"redirecter/pkg/handlers"

	"github.com/gorilla/mux"
)

func main() {
	listenAddr := "0.0.0.0:8090"
	DB := db.Init()
	h := handlers.NewHandler(DB)
	router := mux.NewRouter()
	router.HandleFunc("/links", h.GetAllLinks).Methods(http.MethodGet)
	router.HandleFunc("/links/{id}", h.GetLink).Methods(http.MethodGet)
	router.HandleFunc("/links/{id}", h.UpdateLink).Methods(http.MethodPatch)
	router.HandleFunc("/links/{id}", h.DeleteLink).Methods(http.MethodDelete)
	router.HandleFunc("/links", h.CreateLink).Methods(http.MethodPost)

	router.HandleFunc("/redirects/{uuid}", h.Redirecter)

	router.HandleFunc("/calls/{linkUUID}", h.GetCallsForLink).Methods(http.MethodGet)

	log.Print("Listening on ", listenAddr)
	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("Server exited with error:", err)
	}
}
