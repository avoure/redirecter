package main

import (
	"log"
	"net/http"
	"redirecter/pkg/db"
	"redirecter/pkg/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error while getting .env file. Err: %s", err)
	}
	listenAddr := "0.0.0.0:8090"
	DB := db.Init()
	h := handlers.NewHandler(DB)
	router := mux.NewRouter()

	getAllLinkHandler := handlers.AuthenticateMW(http.HandlerFunc(h.GetAllLinks))
	getLinkHandler := handlers.AuthenticateMW(http.HandlerFunc(h.GetLink))
	updateLinkHandler := handlers.AuthenticateMW(http.HandlerFunc(h.UpdateLink))
	deleteLinkHandler := handlers.AuthenticateMW(http.HandlerFunc(h.DeleteLink))
	createLinkHandler := handlers.AuthenticateMW(http.HandlerFunc(h.DeleteLink))
	getCallsHandler := handlers.AuthenticateMW(http.HandlerFunc(h.GetCallsForLink))

	router.Handle("/links", getAllLinkHandler).Methods(http.MethodGet)
	router.Handle("/links/{id}", getLinkHandler).Methods(http.MethodGet)
	router.Handle("/links/{id}", updateLinkHandler).Methods(http.MethodPatch)
	router.Handle("/links/{id}", deleteLinkHandler).Methods(http.MethodDelete)
	router.Handle("/links", createLinkHandler).Methods(http.MethodPost)

	router.Handle("/calls/{linkUUID}", getCallsHandler).Methods(http.MethodGet)

	router.HandleFunc("/redirects/{uuid}", h.Redirecter)

	log.Print("Listening on ", listenAddr)
	err = http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("Server exited with error:", err)
	}
}
