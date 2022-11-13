package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"redirecter/pkg/models"

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
	json.Unmarshal(requestBody, &link)
	w.Header().Add("Content-Type", "application/json")
	if result := h.DB.Create(&link); result.Error != nil {
		fmt.Println(result.Error)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "Duplicated source url"})
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Created"})
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
