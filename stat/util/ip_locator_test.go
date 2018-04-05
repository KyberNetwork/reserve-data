package util

import "testing"

func TestIPToCountry(t *testing.T) {
	var tests = []struct {
		ip string
		ec string
	}{
		{
			ip: "81.2.69.142",
			ec: "GB",
		},
		{
			ip: "14.177.12.126",
			ec: "VN",
		},
	}

	for _, test := range tests {
		if c, tErr := IPToCountry(test.ip); tErr != nil {
			t.Error(tErr)
		} else {
			if c != test.ec {
				t.Errorf("expected country %q for IP %q, got %q", test.ec, test.ip, c)
			}
		}
	}
}
