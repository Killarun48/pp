package models

import (
	"database/sql"
)

type User struct {
	ID        int          `json:"id" example:"999"`
	Name      string       `json:"name" example:"John Wick"`
	Email     string       `json:"email" example:"wick@continental.com"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty"`
}