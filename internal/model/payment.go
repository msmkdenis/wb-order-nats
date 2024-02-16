package model

type Payment struct {
	Transaction  string `json:"transaction" db:"transaction" validate:"required"`
	RequestID    string `json:"request_id" db:"request_id"`
	Currency     string `json:"currency" db:"currency" validate:"required"`
	Provider     string `json:"provider" db:"provider" validate:"required"`
	Amount       *int   `json:"amount" db:"amount" validate:"gte=0,required"`
	PaymentDt    int64  `json:"payment_dt" db:"payment_dt" validate:"required"`
	Bank         string `json:"bank" db:"bank" validate:"required"`
	DeliveryCost *int   `json:"delivery_cost" db:"delivery_cost" validate:"gte=0,required"`
	GoodsTotal   *int   `json:"goods_total" db:"goods_total" validate:"gte=0,required"`
	CustomFee    *int   `json:"custom_fee" db:"custom_fee" validate:"gte=0,required"`
}
