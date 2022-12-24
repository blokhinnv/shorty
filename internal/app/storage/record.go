package storage

import "time"

type Record struct {
	URL         string    `json:"url"          valid:"url,required"`
	URLID       string    `json:"url_id"       valid:"url,required"`
	UserToken   string    `json:"user_token"`
	Added       time.Time `json:"added"`
	RequestedAt time.Time `json:"requested_at"`
}
