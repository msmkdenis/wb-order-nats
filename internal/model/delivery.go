package model

type Delivery struct {
	Name    string `json:"name" db:"name" validate:"required"`
	Phone   string `json:"phone" db:"phone" validate:"required"`
	Zip     string `json:"zip" db:"zip" validate:"required"`
	City    string `json:"city" db:"city" validate:"required"`
	Address string `json:"address" db:"address" validate:"required"`
	Region  string `json:"region" db:"region" validate:"required"`
	Email   string `json:"email" db:"email" validate:"required,email"`
}
