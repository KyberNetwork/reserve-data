package common

type Feed struct {
	Ask   float64 `json:"ask,string"`
	Bid   float64 `json:"bid,string"`
	Error error   `json:"error,omitempty"`
	Valid bool    `json:"valid"`
}
