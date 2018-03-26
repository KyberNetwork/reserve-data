package util

import (
	"fmt"
	"log"
	"net"
	"errors"

	"github.com/oschwald/geoip2-golang"
)

const GEO_DATABASE string = "/go/src/github.com/KyberNetwork/reserve-data/stat/util/GeoLite2-Country.mmdb"

var IPLocator *geoip2.Reader

func init() {
	var err error
	IPLocator, err = geoip2.Open(GEO_DATABASE)
	if err != nil {
		panic(err)
	}
}

func IpToCountry(IP string) (string, error) {
	var country string

	IPParsed := net.ParseIP(IP)
	if IPParsed == nil {
		return country, errors.New(fmt.Sprintf("IP %s is not valid!", IP))
	}
	record, err := IPLocator.Country(IPParsed)
	if err != nil {
		log.Printf("failed to query data from geo-database!")
		return country, err
	}
	
	country = record.Country.Names["en"] //name of country in english
	if country == "" {
		return country, errors.New(fmt.Sprintf("Can't find country of the given IP: %s", IP))
	}
	return  country, nil
}
