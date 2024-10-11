package controller

type RequestUser struct {
	Name  string `json:"name" example:"John Wick"`
	Email string `json:"email" example:"wick@continental.com"`
}