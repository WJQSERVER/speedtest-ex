package ipinfo

import (
	"fmt"
	"regexp"
	"speedtest/config"
	"speedtest/results"
)

// 预编译的正则表达式变量
var (
	localIPv6Regex          = regexp.MustCompile(`^::1$`)                            // 匹配本地 IPv6 地址
	linkLocalIPv6Regex      = regexp.MustCompile(`^fe80:`)                           // 匹配链路本地 IPv6 地址
	localIPv4Regex          = regexp.MustCompile(`^127\.`)                           // 匹配本地 IPv4 地址
	privateIPv4Regex10      = regexp.MustCompile(`^10\.`)                            // 匹配私有 IPv4 地址（10.0.0.0/8）
	privateIPv4Regex172     = regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`)       // 匹配私有 IPv4 地址（172.16.0.0/12）
	privateIPv4Regex192     = regexp.MustCompile(`^192\.168\.`)                      // 匹配私有 IPv4 地址（192.168.0.0/16）
	linkLocalIPv4Regex      = regexp.MustCompile(`^169\.254\.`)                      // 匹配链路本地 IPv4 地址（169.254.0.0/16）
	cgnatIPv4Regex          = regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`) // 匹配 CGNAT IPv4 地址（100.64.0.0/10）
	unspecifiedAddressRegex = regexp.MustCompile(`^0\.0\.0\.0$`)                     // 匹配未指定地址（0.0.0.0）
	broadcastAddressRegex   = regexp.MustCompile(`^255\.255\.255\.255$`)             // 匹配广播地址（255.255.255.255）
	removeASRegexp          = regexp.MustCompile(`AS\d+\s`)                          // 用于去除 ISP 信息中的自治系统编号
)

func GetIP(clientIP string, cfg *config.Config) (results.Result, error) {
	var ret results.Result // 创建结果结构体实例
	// 使用正则表达式匹配不同类型的 IP 地址
	switch {
	case localIPv6Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - localhost IPv6 access" // 本地 IPv6 地址
	case linkLocalIPv6Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - link-local IPv6 access" // 链路本地 IPv6 地址
	case localIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - localhost IPv4 access" // 本地 IPv4 地址
	case privateIPv4Regex10.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（10.0.0.0/8）
	case privateIPv4Regex172.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（172.16.0.0/12）
	case privateIPv4Regex192.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access" // 私有 IPv4 地址（192.168.0.0/16）
	case linkLocalIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - link-local IPv4 access" // 链路本地 IPv4 地址
	case cgnatIPv4Regex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access" // CGNAT IPv4 地址（100.64.0.0/10）
	case unspecifiedAddressRegex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - unspecified address" // 未指定地址（0.0.0.0）
	case broadcastAddressRegex.MatchString(clientIP):
		ret.ProcessedString = clientIP + " - broadcast address" // 广播地址（255.255.255.255）
	default:
		ret.ProcessedString = clientIP // 其他情况，返回原始 IP 地址
	}
	/*
		// 检查处理结果中是否包含特定信息
		if strings.Contains(ret.ProcessedString, " - ") {
			// 将 ret 转换为 JSON 字符串
			jsonData, err := json.Marshal(ret)
			if err != nil {
				// 如果转换失败，记录错误信息
				logInfo("Error marshaling JSON: " + err.Error())
			} else {
				// 如果转换成功，记录 JSON 字符串
				logInfo(string(jsonData))
			}
			return ret // 返回结果
		} */

	ispInfo, err := getIPInfo(clientIP, cfg)
	if err != nil {
		return results.Result{}, err
	}
	//ret.RawISPInfo = ispInfo // 存储原始 ISP 信息
	// 转写 ISP 信息
	ret.RawISPInfo = results.CommonIPInfoResponse{
		IP:        ispInfo.IP,
		Org:       ispInfo.Org,
		Region:    ispInfo.Region,
		City:      ispInfo.City,
		Country:   ispInfo.Country,
		Continent: ispInfo.Continent,
	}
	/*


		isp := removeASRegexp.ReplaceAllString(ispInfo.ISP, "") // 去除 ISP 信息中的自治系统编号

		if isp == "" {
			isp = "Unknown ISP" // 如果 ISP 信息为空，设置为未知
		}

		if ispInfo.CountryName != "" {
			isp += ", " + ispInfo.CountryName // 如果有国家名称，添加到 ISP 信息中
		}

		ret.ProcessedString += " - " + isp // 更新处理后的字符串
	*/
	ret.ProcessedString = MakeProcessedString(ret.ProcessedString, ispInfo) // 更新处理后的字符串
	return ret, nil                                                         // 返回结果
}

// 获取 IP 地址信息
func getIPInfo(ip string, cfg *config.Config) (CommonIPInfoResponse, error) {

	switch cfg.IPinfo.Model {
	case "ip":
		// 自托管 IP 信息查询
		var ret CommonIPInfoResponse // 创建结果结构体实例
		ret, err := getHostIPInfo(ip, cfg)
		if err != nil {
			return CommonIPInfoResponse{}, err
		}
		return ret, nil
	case "ipinfo":
		// ipinfo.io 信息查询
		var ret CommonIPInfoResponse // 创建结果结构体实例
		ret, err := getIPInfoIO(ip, cfg)
		if err != nil {
			return CommonIPInfoResponse{}, err
		}
		return ret, nil
	default:
		// 模型不支持
		return CommonIPInfoResponse{}, fmt.Errorf("Unsupported IPinfo model: " + cfg.IPinfo.Model)
	}
}

func MakeProcessedString(processedString string, ispInfo CommonIPInfoResponse) string {
	info := processedString
	if ispInfo.Org != "" {
		info += " - " + ispInfo.Org
	}
	if ispInfo.Region != "" {
		info += " - " + ispInfo.Region
	}
	if ispInfo.City != "" {
		info += " - " + ispInfo.City
	}
	if ispInfo.Country != "" {
		info += " - " + ispInfo.Country
	}
	if ispInfo.Continent != "" {
		info += " - " + ispInfo.Continent
	}
	return info
}
