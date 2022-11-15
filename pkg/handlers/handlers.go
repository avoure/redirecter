package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"redirecter/pkg/models"

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
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var link models.RedirectMap
	link.UUID = uuid.New()
	json.Unmarshal(requestBody, &link)
	u, err := url.ParseRequestURI(*link.DestinationURL)
	if err != nil {
		fmt.Println("Requested destinationUrl: ", u)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "Invalid destination URL"})
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if result := h.DB.Create(&link); result.Error != nil {
		fmt.Println(result.Error)
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var updatedLink models.RedirectMap
	json.Unmarshal(body, &updatedLink)
	var link models.RedirectMap
	if result := h.DB.First(&link, linkId); result.Error != nil {
		fmt.Println(result.Error)
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
		fmt.Println(result.Error)
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
		fmt.Println(result.Error)
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
		fmt.Println(result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Link not found"})

	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(link)
	}
}

func (h Handler) Redirecter(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	var link models.RedirectMap
	if result := h.DB.Find(&link, "UUID = ?", uuid); result.Error != nil {
		fmt.Println(result.Error)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Cannot redirect"})
	} else {
		http.Redirect(w, r, *link.DestinationURL, http.StatusFound)
	}
}
