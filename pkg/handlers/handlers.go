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
	if err := json.Unmarshal(requestBody, &link); err != nil {
		log.Println("Cannot unpack request body")
	}
	if link.DestinationURL != "" {
		_, err = url.ParseRequestURI(link.DestinationURL)
	}
	if err != nil || link.DestinationURL == "" {
		log.Printf("Invalid destinationUrl: %s : %s", link.DestinationURL, err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Invalid destination URL"}); err != nil {
			log.Println("Cannot write response")
		}
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if result := h.DB.Create(&link); result.Error != nil {
		log.Printf("Failed to create link: %s", result.Error)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Duplicated source url"}); err != nil {
			log.Println("Cannot write response")
		}
	} else {
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(link); err != nil {
			log.Println("Cannot write response")
		}
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
	if err := json.Unmarshal(body, &updatedLink); err != nil {
		log.Println("Cannot unpack request body")
	}
	var link models.RedirectMap
	if result := h.DB.First(&link, linkId); result.Error != nil {
		log.Printf("Failed to update link %s: %s", linkId, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"}); err != nil {
			log.Println("Cannot unpack request body")
		}
		return
	}
	link.DestinationURL = updatedLink.DestinationURL
	h.DB.Save(&link)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(link); err != nil {
		log.Println("Cannot unpack request body")
	}
}

func (h Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	linkId := mux.Vars(r)["id"]
	var link models.RedirectMap
	if result := h.DB.First(&link, linkId); result.Error != nil {
		log.Printf("Failed to delete link %s: %s", linkId, result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"}); err != nil {
			log.Println("Cannot unpack request body")
		}
	} else {
		h.DB.Delete(&link)
		w.WriteHeader(http.StatusAccepted)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Link successfully deleted"}); err != nil {
			log.Println("Cannot unpack request body")
		}
	}

}

func (h Handler) GetAllLinks(w http.ResponseWriter, r *http.Request) {
	var links []models.RedirectMap

	if result := h.DB.Find(&links); result.Error != nil {
		log.Printf("Failed to get links: %s", result.Error)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		log.Println("Cannot unpack request body")
	}
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
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"}); err != nil {
			log.Println("Cannot unpack request body")
		}

	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(link); err != nil {
			log.Println("Cannot unpack request body")
		}
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
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "Cannot redirect"}); err != nil {
			log.Println("Cannot unpack request body")
		}
	}

	http.Redirect(w, r, link.DestinationURL, http.StatusTemporaryRedirect)

	log.Printf("Redirected call %s for %s", callID, linkUUID)

	body := []byte{}
	defer r.Body.Close()
	if bodyBytes, err := io.ReadAll(r.Body); err != nil {
		log.Printf("Failed to read body for call %s: %s", callID, err)
	} else {
		body = bodyBytes
	}
	go h.storeCall(callID, link, r, body)
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
	if err := json.NewEncoder(w).Encode(calls); err != nil {
		log.Println("Cannot unpack request body")
	}
}

func (h Handler) storeCall(callID uuid.UUID, link models.RedirectMap, r *http.Request, body []byte) {
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

	log.Printf("Logged call %s for link %s in DB.", newCall.ID, link.UUID)
}
