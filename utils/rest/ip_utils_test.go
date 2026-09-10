package rest

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := &http.Request{
		Header:     http.Header{HeaderXForwardedFor: []string{"198.51.100.1, 203.0.113.45"}},
		RemoteAddr: "203.0.113.99:12345",
	}
	ip := GetClientIP(req)
	assert.Equal(t, "198.51.100.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := &http.Request{
		Header: http.Header{
			HeaderXRealIP: []string{"198.51.100.17"},
		},
		RemoteAddr: "203.0.113.99:1234", // fallback, should NOT be used
	}

	ip := GetClientIP(req)

	assert.Equal(t, "198.51.100.17", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := &http.Request{
		Header:     http.Header{},
		RemoteAddr: "203.0.113.99:12345",
	}
	ip := GetClientIP(req)
	assert.Equal(t, "203.0.113.99", ip)
}

func TestGetClientIP_BadRemoteAddr(t *testing.T) {
	req := &http.Request{
		Header:     http.Header{},
		RemoteAddr: "not-an-ip",
	}
	ip := GetClientIP(req)
	assert.Equal(t, "not-an-ip", ip)
}

func TestIsPrivateIP(t *testing.T) {
	assert.True(t, IsPrivateIP("10.0.0.1"))
	assert.True(t, IsPrivateIP("172.16.5.1"))
	assert.True(t, IsPrivateIP("192.168.1.1"))
	assert.True(t, IsPrivateIP("fc00::1"))
	assert.False(t, IsPrivateIP("8.8.8.8"))
	assert.False(t, IsPrivateIP("invalid-ip"))
}

func TestNormalizeIP(t *testing.T) {
	assert.Equal(t, "192.168.1.1", NormalizeIP(" 192.168.1.1 "))
	assert.Equal(t, "", NormalizeIP("not-an-ip"))
	assert.Equal(t, "::1", NormalizeIP(" ::1 "))
}

func TestIsValidIP(t *testing.T) {
	assert.True(t, IsValidIP("127.0.0.1"))
	assert.True(t, IsValidIP("::1"))
	assert.False(t, IsValidIP("invalid-ip"))
}

func TestGetIPVersion(t *testing.T) {
	assert.Equal(t, 4, GetIPVersion("192.168.1.1"))
	assert.Equal(t, 6, GetIPVersion("::1"))
	assert.Equal(t, 0, GetIPVersion("not-an-ip"))
}

func TestMaskIP(t *testing.T) {
	assert.Equal(t, "192.168.*.*", MaskIP("192.168.1.100"))
	assert.Equal(t, "2001:db8::*", MaskIP("2001:db8::ff00:42:8329"))
	assert.Equal(t, "", MaskIP("invalid-ip"))
}

func TestMaskIP_IPv6NilBranch(t *testing.T) {
	old := parseIP
	parseIP = func(string) net.IP {
		return net.IP{}
	}
	defer func() {
		parseIP = old
	}()

	assert.Equal(t, "", MaskIP("anything"))
}
