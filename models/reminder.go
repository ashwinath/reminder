package models

import "time"

type Reminder struct {
	ID          int64
	Status      string
	Description string
	URL         string
	Remarks     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
