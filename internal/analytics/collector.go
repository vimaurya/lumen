package analytics

import (
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/ua-parser/uap-go/uaparser"
)

var HitBuffer = make(chan Hit, 5000)

var (
	dropCounter  = 0
	GeoDB        *geoip2.Reader
	ClientParser *uaparser.Parser
)

func Collect(hit Hit) {
	select {
	case HitBuffer <- hit:
	default:
		dropCounter += 1
	}
}

func extractIP(ip string) string {
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}

	return host
}

func getCountry(ip string) string {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return "Unknown"
	}

	record, err := GeoDB.City(parsedIP)
	if err != nil {
		return "Unknown"
	}

	if name, ok := record.Country.Names["en"]; ok {
		return name
	}

	return record.Country.IsoCode
}
