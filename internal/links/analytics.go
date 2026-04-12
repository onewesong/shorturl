package links

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const analyticsRecentVisitLimit = 20

func normalizeAnalyticsDays(days int) int {
	switch {
	case days <= 0:
		return 7
	case days > 90:
		return 90
	default:
		return days
	}
}

func maskIP(raw string) string {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return ""
	}

	if v4 := ip.To4(); v4 != nil {
		return strconv.Itoa(int(v4[0])) + "." + strconv.Itoa(int(v4[1])) + "." + strconv.Itoa(int(v4[2])) + ".*"
	}

	parts := strings.Split(ip.String(), ":")
	if len(parts) <= 2 {
		return ip.String()
	}

	return strings.Join(parts[:2], ":") + ":*"
}

func MaskVisitIP(raw string) string {
	return maskIP(raw)
}

func extractRefererHost(raw string) string {
	referer := strings.TrimSpace(raw)
	if referer == "" {
		return ""
	}

	parsed, err := url.Parse(referer)
	if err != nil {
		return ""
	}

	return parsed.Hostname()
}

func detectClient(userAgent string) (clientName string, clientType string, deviceType string, osName string) {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return "未知客户端", "unknown", "unknown", "未知系统"
	}

	deviceType = "desktop"
	switch {
	case strings.Contains(ua, "ipad"), strings.Contains(ua, "tablet"):
		deviceType = "tablet"
	case strings.Contains(ua, "mobile"), strings.Contains(ua, "iphone"), strings.Contains(ua, "android"):
		deviceType = "mobile"
	}

	switch {
	case strings.Contains(ua, "windows"):
		osName = "Windows"
	case strings.Contains(ua, "android"):
		osName = "Android"
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"), strings.Contains(ua, "ios"):
		osName = "iOS"
	case strings.Contains(ua, "mac os x"), strings.Contains(ua, "macintosh"):
		osName = "macOS"
	case strings.Contains(ua, "linux"):
		osName = "Linux"
	default:
		osName = "未知系统"
	}

	switch {
	case strings.Contains(ua, "bot"), strings.Contains(ua, "spider"), strings.Contains(ua, "crawler"):
		return "Bot", "bot", "bot", osName
	case strings.Contains(ua, "micromessenger"):
		return "微信", "app", deviceType, osName
	case strings.Contains(ua, "weibo"):
		return "微博", "app", deviceType, osName
	case strings.Contains(ua, "douyin"):
		return "抖音", "app", deviceType, osName
	case strings.Contains(ua, "qq/"):
		return "QQ", "app", deviceType, osName
	case strings.Contains(ua, "postman"):
		return "Postman", "tool", deviceType, osName
	case strings.Contains(ua, "curl"):
		return "curl", "tool", deviceType, osName
	case strings.Contains(ua, "edg/"):
		return "Edge", "browser", deviceType, osName
	case strings.Contains(ua, "opr/"), strings.Contains(ua, "opera"):
		return "Opera", "browser", deviceType, osName
	case strings.Contains(ua, "firefox/"):
		return "Firefox", "browser", deviceType, osName
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/") && !strings.Contains(ua, "opr/"):
		return "Chrome", "browser", deviceType, osName
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/"):
		return "Safari", "browser", deviceType, osName
	default:
		return "未知客户端", "unknown", deviceType, osName
	}
}

func normalizeVisitMeta(meta VisitMeta) VisitMeta {
	if meta.VisitedAt.IsZero() {
		meta.VisitedAt = time.Now().UTC()
	}
	meta.Referer = strings.TrimSpace(meta.Referer)
	if meta.RefererHost == "" {
		meta.RefererHost = extractRefererHost(meta.Referer)
	}
	meta.UserAgent = strings.TrimSpace(meta.UserAgent)
	if meta.ClientName == "" || meta.ClientType == "" || meta.DeviceType == "" || meta.OS == "" {
		meta.ClientName, meta.ClientType, meta.DeviceType, meta.OS = detectClient(meta.UserAgent)
	}

	return meta
}
