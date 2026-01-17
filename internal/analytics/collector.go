package analytics

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
	"time"

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

func getSessionId(ip string, ua string) string {
	window := time.Now().Unix() / 1800
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", ip, ua, window)))
	return fmt.Sprintf("%x", hash)[:16]
}

func isBot(ua string) bool {
	client := ClientParser.Parse(ua)
	isBot := strings.Contains(strings.ToLower(ua), "bot") ||
		strings.Contains(strings.ToLower(ua), "crawler") ||
		client.Device.Family == "Spider"

	return isBot
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
