package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	var x struct {
		Duration HumanDuration `json:"d"`
	}
	err := json.Unmarshal([]byte(`{"d":"10s"}`), &x)
	assert.NoError(t, err)
	assert.Equal(t, x.Duration, HumanDuration(time.Second*10))
}
