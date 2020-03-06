package common

import "encoding/json"

type RespError struct {
	Msg        string
	Body       interface{}
	StatusCode int
}

func (r *RespError) Error() string {
	b, err := json.Marshal(r)
	if err != nil {
		return r.Msg
	}
	return string(b)
}
