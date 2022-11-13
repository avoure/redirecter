package models

import "time"

type RedirectMap struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	SourceURL      *string   `json:"sourceUrl" gorm:"unique"`
	DestinationURL *string   `json:"destinationUrl"`
}
