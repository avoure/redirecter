package models

import (
	"time"

	"github.com/google/uuid"
)

type RedirectMap struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UUID           uuid.UUID `json:"uuid"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	DestinationURL *string   `json:"destinationUrl" gorm:"unique"`
}

type IncomingCall struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time `json:"createdAt"`
	RedirectUUID uuid.UUID `json:"redirectUUID"`
	Method       string    `json:"method"`
	Headers      string    `json:"headers"`
	Body         []byte    `json:"body"`
}
