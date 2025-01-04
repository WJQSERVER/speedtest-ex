package ipinfo

import (
	"encoding/json"
	"io"
	"speedtest/config"

	"github.com/imroc/req/v3"
)

/*
// IPInfoResponse format (ipinfo.io version)
type IPInfoResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
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
*/

func getIPInfoURL(ip string, apiKey string) string {
	if apiKey == "" {
		return "https://ipinfo.io/" + ip + "/json"
	} else {
		return "https://ipinfo.io/" + ip + "/json?token=" + apiKey
	}
}

func getIPInfoIO(ip string, cfg *config.Config) (CommonIPInfoResponse, error) {
	selfhostApi := getIPInfoURL(ip, cfg.IPinfo.IPinfoKey)
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

	// 使用IPInfoResponse结构体解析json数据
	var ipHost IPInfoResponse
	err = json.Unmarshal(bodyBytes, &ipHost)
	if err != nil {
		return CommonIPInfoResponse{}, err
	}

	// 通过现有IPHostResponse结构体制作CommonIPInfoResponse结构体
	commonIPInfo := CommonIPInfoResponse{
		IP:        ip,
		Org:       ipHost.Org,
		Region:    ipHost.Region,
		City:      ipHost.City,
		Continent: "",
	}
	return commonIPInfo, nil

}
