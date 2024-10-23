package models

import (
	"database/sql"
	"encoding/json"
)

type User struct {
	ID        int          `json:"id" example:"999"`
	Name      string       `json:"name" example:"John Wick"`
	Email     string       `json:"email" example:"wick@continental.com"`
	DeletedAt sql.NullTime `json:"deleted_at,omitempty" swaggertype:"string" example:"2022-02-01T00:00:00Z"`
}

func (u User) MarshalJSON() ([]byte, error) {
	response := struct {
		ID        int    `json:"id" example:"999"`
		Name      string `json:"name" example:"John Wick"`
		Email     string `json:"email" example:"wick@continental.com"`
		DeletedAt string `json:"deleted_at,omitempty" swaggertype:"string" example:"2022-02-01T00:00:00Z"`
	}{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
	}

	if u.DeletedAt.Valid {
		response.DeletedAt = u.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	jsonResp, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonResp, nil
}
