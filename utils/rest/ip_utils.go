package rest

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var parseIP = net.ParseIP

// GetClientIP extracts the real client IP address from the HTTP request.
// It checks X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
func GetClientIP(r *http.Request) string {
	xff := r.Header.Get(HeaderXForwardedFor) // Check for X-Forwarded-For header first
	if xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if ip != "" {
				return ip
			}
		}
	}

	// Then check X-Real-IP
	if xrip := strings.TrimSpace(r.Header.Get(HeaderXRealIP)); xrip != "" {
		return xrip
	}

	// Fallback to RemoteAddr (remove port if present)
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr) // return full addr if unable to split
	}

	return ip
}

// IsPrivateIP checks if the given IP address belongs to a private or reserved IP range.
// It supports both IPv4 (e.g., 10.0.0.0/8, 192.168.0.0/16) and IPv6 (e.g., fc00::/7) blocks.
//
// Returns true if the IP is private, false otherwise.
// Returns false for invalid or unparseable IPs.
func IsPrivateIP(ipStr string) bool {
	ip := parseIP(ipStr)
	if ip == nil {
		return false
	}

	privateBlocks := []*net.IPNet{
		// IPv4 private blocks
		mustCIDR("10.0.0.0/8"),
		mustCIDR("172.16.0.0/12"),
		mustCIDR("192.168.0.0/16"),

		// IPv6 unique local address block
		mustCIDR("fc00::/7"),
	}

	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// mustCIDR is a helper function that parses a CIDR string and returns the *net.IPNet.
// It panics if the CIDR string is invalid, and is typically used for hardcoded CIDRs.
func mustCIDR(cidr string) *net.IPNet {
	_, block, _ := net.ParseCIDR(cidr)
	return block
}

// NormalizeIP trims spaces and normalizes the IP address string.
// It parses the input string and returns its canonical form.
// If the input is invalid, it returns an empty string.
//
// Useful for cleaning up IP inputs before processing or logging.
func NormalizeIP(ip string) string {
	parsed := parseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ""
	}
	return parsed.String()
}

// IsValidIP checks whether the provided string is a valid IP address.
// It returns true for valid IPv4 or IPv6 addresses, and false otherwise.
func IsValidIP(ipStr string) bool {
	return parseIP(ipStr) != nil
}

// GetIPVersion determines the IP version of the given address string.
// Returns 4 for IPv4, 6 for IPv6, and 0 if the IP is invalid or unrecognized.
func GetIPVersion(ipStr string) int {
	ip := parseIP(ipStr)
	if ip == nil {
		return 0
	}
	if ip.To4() != nil {
		return 4
	}
	return 6
}

// MaskIP returns a partially masked (anonymized) version of the IP address.
// For IPv4, the last two octets are replaced with asterisks (e.g., 192.168.*.*).
// For IPv6, all but the first two segments are masked (e.g., abcd:1234::*).
// If the IP is invalid, it returns an empty string.
//
// This is useful for logging or displaying user IPs without revealing the full address.
func MaskIP(ipStr string) string {
	ip := parseIP(ipStr)
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return fmt.Sprintf("%d.%d.*.*", ipv4[0], ipv4[1])
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return ""
	}
	return fmt.Sprintf("%x:%x::*", uint16(ipv6[0])<<8|uint16(ipv6[1]), uint16(ipv6[2])<<8|uint16(ipv6[3]))
}
