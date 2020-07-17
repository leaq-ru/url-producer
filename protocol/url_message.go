package protocol

import "time"

type URLMessage struct {
	URL              string    `json:"url"`
	Registrar        string    `json:"registrar"`
	RegistrationDate time.Time `json:"registration_date"`
}
