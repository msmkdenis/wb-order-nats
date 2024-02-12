package model

type Item struct {
	ChrtID      int    `json:"chrt_id" db:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" db:"track_number" validate:"required"`
	Price       int    `json:"price" db:"price" validate:"required"`
	Rid         string `json:"rid" db:"rid" validate:"required"`
	Name        string `json:"name" db:"name" validate:"required"`
	Sale        int    `json:"sale" db:"sale" validate:"required"`
	Size        string `json:"size" db:"size" validate:"required"`
	TotalPrice  int    `json:"total_price,omitempty" db:"total_price"`
	NmID        int    `json:"nm_id" db:"nm_id" validate:"required"`
	Brand       string `json:"brand" db:"brand" validate:"required"`
	Status      int    `json:"status" db:"status" validate:"required"`
}
