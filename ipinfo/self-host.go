package ipinfo

import (
	"encoding/json"
	"io"
	"speedtest/config"

	"github.com/imroc/req/v3"
)

/*
// IPHostResponse format (self-host version)
type IPHostResponse struct {
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
	Rrgion    string `json:"region"`    // ipinfo = region, self-host = nil
	City      string `json:"city"`      // ipinfo = city, self-host = nil
	Country   string `json:"country"`   // ipinfo = Country, self-host = CountryCode
	Continent string `json:"continent"` // ipinfo = nil, self-host = continent_name
}
*/

func getHostIPInfo(ip string, cfg *config.Config) (CommonIPInfoResponse, error) {
	selfhostApi := cfg.IPinfo.IPinfoURL + "/api/ip-lookup?ip=" + ip
	// 使用req库发送请求并使用chrome的TLS指纹
	client := req.C().
		SetTLSFingerprintChrome().
		ImpersonateChrome()

	resp, err := client.R().Get(selfhostApi)
	if err != nil {
		return CommonIPInfoResponse{}, err
	}
	defer resp.Body.Close()

	// 读取body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {

		return CommonIPInfoResponse{}, err
	}
	logInfo("self-host response: " + string(bodyBytes))

	// 使用IPHostResponse结构体解析json数据
	var ipHost IPHostResponse
	err = json.Unmarshal(bodyBytes, &ipHost)
	if err != nil {
		return CommonIPInfoResponse{}, err
	}

	// 通过现有IPHostResponse结构体制作CommonIPInfoResponse结构体
	commonIPInfo := CommonIPInfoResponse{
		IP:        ip,
		Org:       ipHost.ISP + " " + ipHost.ASN,
		Region:    "",
		City:      "",
		Country:   ipHost.CountryCode,
		Continent: ipHost.ContinentCode,
	}
	return commonIPInfo, nil
}
