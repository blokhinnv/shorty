package storage

import "time"

// Record структура для хранения записи из хранилища.
type Record struct {
	URL         string    `json:"url"          valid:"url,required"`
	URLID       string    `json:"url_id"       valid:"url,required"`
	UserID      uint32    `json:"user_id"`
	Added       time.Time `json:"added"`
	RequestedAt time.Time `json:"requested_at"`
	IsDeleted   bool      `json:"is_deleted"`
}
