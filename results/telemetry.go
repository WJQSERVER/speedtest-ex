package results

import (
	_ "embed"
	"math/rand"
	"net"
	"net/http"
	"time"

	"speedtest/config"
	"speedtest/database"
	"speedtest/database/schema"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

type Result struct {
	ProcessedString string               `json:"processedString"`
	RawISPInfo      CommonIPInfoResponse `json:"rawIspInfo"`
}

// IPInfoResponse format (self-host version)
type IPInfoResponse struct {
	IP            string `json:"ip"`             // IP address (IPv4 or IPv6)
	ASN           string `json:"asn"`            // Autonomous System Number
	Domain        string `json:"domain"`         // Domain name
	ISP           string `json:"isp"`            // Internet Service Provider
	ContinentCode string `json:"continent_code"` // Continent code
	ContinentName string `json:"continent_name"` // Continent name
	CountryCode   string `json:"country_code"`
	CountryName   string `json:"country_name"`
	UserAgent     string `json:"user_agent"`
}

// 通用响应格式
type CommonIPInfoResponse struct {
	IP        string `json:"ip"`
	Org       string `json:"org"`       // ipinfo = org, self-host = ASN + ISP
	Region    string `json:"region"`    // ipinfo = region, self-host = nil
	City      string `json:"city"`      // ipinfo = city, self-host = nil
	Country   string `json:"country"`   // ipinfo = Country, self-host = CountryCode
	Continent string `json:"continent"` // ipinfo = nil, self-host = continent_name
}

func Record(c *gin.Context, cfg *config.Config) {
	if cfg.Database.Model == "none" {
		c.String(http.StatusOK, "Telemetry is disabled")
		return
	}

	ipAddr, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	userAgent := c.Request.UserAgent()
	language := c.Request.Header.Get("Accept-Language")

	ispInfo := c.PostForm("ispinfo")
	//logInfo("debug > result > ispInfo: %s", ispInfo)
	download := c.PostForm("dl")
	upload := c.PostForm("ul")
	ping := c.PostForm("ping")
	jitter := c.PostForm("jitter")
	logs := c.PostForm("log")
	extra := c.PostForm("extra")

	var record schema.TelemetryData
	record.IPAddress = ipAddr
	if ispInfo == "" {
		record.ISPInfo = "{}"
	} else {
		record.ISPInfo = ispInfo
	}
	record.Extra = extra
	record.UserAgent = userAgent
	record.Language = language
	record.Download = download
	record.Upload = upload
	record.Ping = ping
	record.Jitter = jitter
	record.Log = logs

	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	uuid := ulid.MustNew(ulid.Timestamp(t), entropy)
	record.UUID = uuid.String()

	err := database.DB.SaveTelemetry(&record)
	if err != nil {
		logError("Error inserting into database: %s", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusOK, "id "+uuid.String())
}
