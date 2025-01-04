package ipinfo

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
