package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"redirecter/pkg/models"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) Handler {
	return Handler{db}
}

func (h Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var link models.RedirectMap
	link.UUID = uuid.New()
	json.Unmarshal(requestBody, &link)
	if link.DestinationURL != "" {
		_, err = url.ParseRequestURI(link.DestinationURL)
	}
	if err != nil || link.DestinationURL == "" {
		log.Printf("Invalid destinationUrl: %s : %s", link.DestinationURL, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "Invalid destination URL"})
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if result := h.DB.Create(&link); result.Error != nil {
		log.Printf("Failed to create link: %s", result.Error)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "Duplicated source url"})
	} else {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(link)
	}
}

func (h Handler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	linkId := mux.Vars(r)["id"]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var updatedLink models.RedirectMap
	json.Unmarshal(body, &updatedLink)
	var link models.RedirectMap
	if result := h.DB.First(&link, linkId); result.Error != nil {
		log.Printf("Failed to update link %s: %s", linkId, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"})
		return
	}
	link.DestinationURL = updatedLink.DestinationURL
	h.DB.Save(&link)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(link)
}

func (h Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	linkId := mux.Vars(r)["id"]
	var link models.RedirectMap
	if result := h.DB.First(&link, linkId); result.Error != nil {
		log.Printf("Failed to delete link %s: %s", linkId, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"})
	} else {
		h.DB.Delete(&link)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "Link successfully deleted"})
	}

}

func (h Handler) GetAllLinks(w http.ResponseWriter, r *http.Request) {
	var links []models.RedirectMap

	if result := h.DB.Find(&links); result.Error != nil {
		log.Printf("Failed to get links: %s", result.Error)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(links)
}

func (h Handler) GetLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId := vars["id"]
	var link models.RedirectMap
	w.Header().Add("Content-Type", "application/json")
	if result := h.DB.First(&link, linkId); result.Error != nil {
		log.Printf("Failed to get link %s: %s", linkId, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"})

	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(link)
	}
}

func (h Handler) Redirecter(w http.ResponseWriter, r *http.Request) {
	callID := uuid.New()
	linkUUID := mux.Vars(r)["uuid"]
	var link models.RedirectMap
	if result := h.DB.Find(&link, "UUID = ?", linkUUID); result.Error != nil {
		log.Printf("Failed to get link %s: %s", linkUUID, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Cannot redirect"})
	}

	http.Redirect(w, r, link.DestinationURL, http.StatusFound)

	log.Printf("Redirected call %s for %s, logging attempt.", callID, linkUUID)

	headers := ""
	if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
		log.Printf("Failed to Marshal request Headers for call %s: %s", callID, err)
	} else {
		headers = string(reqHeadersBytes)
	}

	queryParams := ""
	if qpBytes, err := json.Marshal(r.URL.Query()); err != nil {
		log.Printf("Failed to Marshal request query params for call %s: %s", callID, err)
	} else {
		queryParams = string(qpBytes)
	}

	body := []byte{}
	defer r.Body.Close()
	if bodyBytes, err := io.ReadAll(r.Body); err != nil {
		log.Printf("Failed to read body for call %s: %s", callID, err)
	} else {
		body = bodyBytes
	}

	newCall := models.IncomingCall{
		ID:           callID,
		CreatedAt:    time.Now(),
		RedirectUUID: link.UUID,
		Method:       r.Method,
		Headers:      headers,
		QueryParams:  queryParams,
		Body:         body,
	}
	if result := h.DB.Create(&newCall); result.Error != nil {
		log.Printf("Failed to store call %s in DB: %s", callID, result.Error)
	}

	log.Printf("Logged attempt %s for link %s in DB.", newCall.ID, link.UUID)
}

func (h Handler) GetCallsForLink(w http.ResponseWriter, r *http.Request) {
	var calls []models.IncomingCall
	linkUUID := mux.Vars(r)["linkUUID"]
	if result := h.DB.Order("created_at desc").Find(&calls, "redirect_uuid = ?", linkUUID); result.Error != nil {
		log.Printf("Failed to retrieve calls for %s: %s", linkUUID, result.Error)

		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(calls)
}
