package income

import "time"

type Income struct {
	Id        int       `json:"id"`
	Date      time.Time `json:"date"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	Note      string    `json:"note"`
	ImageUrl  string    `json:"image_url"`
	SpenderId int       `json:"spender_id"`
}
