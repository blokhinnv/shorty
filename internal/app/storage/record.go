package storage

import "time"

type Record struct {
	URL         string    `json:"url"          valid:"url,required"`
	URLID       string    `json:"url_id"       valid:"url,required"`
	UserID      string    `json:"user_id"`
	Added       time.Time `json:"added"`
	RequestedAt time.Time `json:"requested_at"`
}
