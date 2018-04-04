package util

import (
	"fmt"
	"log"
	"net"
	"path"

	"github.com/oschwald/geoip2-golang"

	"github.com/KyberNetwork/reserve-data/common"
)

// IPToCountry initializes a new ipLocator and resolve the given IP address.
func IPToCountry(ip string) (string, error) {
	il, err := newIPLocator()
	if err != nil {
		return "", err
	}
	return il.ipToCountry(ip)
}

// ipLocator is a resolver that query data of IP from MaxMind's GeoLite2 database.
type ipLocator struct {
	r *geoip2.Reader
}

// newIPLocator returns an instance of ipLocator.
func newIPLocator() (*ipLocator, error) {
	dbPath := path.Join(common.CurrentDir(), "GeoLite2-Country.mmdb")
	r, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &ipLocator{r: r}, nil
}

// ipToCountry returns the country of given IP address.
func (il *ipLocator) ipToCountry(ip string) (string, error) {
	IPParsed := net.ParseIP(ip)
	if IPParsed == nil {
		return "", fmt.Errorf("ip %s is not valid!", ip)
	}
	record, err := il.r.Country(IPParsed)
	if err != nil {
		log.Printf("failed to query data from geo-database!")
		return "", err
	}

	country := record.Country.IsoCode //iso code of country
	if country == "" {
		return "", fmt.Errorf("Can't find country of the given ip: %s", ip)
	}
	return country, nil
}
