package system

import (
	"fmt"
	"net"
	"testing"
)

func TestGetInMemorySizeInBytes_UnsupportedInSliceAndMap(t *testing.T) {
	_, err := GetInMemorySizeInBytes([]interface{}{make(chan int)})
	if err == nil {
		t.Fatalf("expected error for unsupported slice element")
	}

	_, err = GetInMemorySizeInBytes(map[string]interface{}{"x": make(chan int)})
	if err == nil {
		t.Fatalf("expected error for unsupported map value")
	}
}

func TestGetInMemorySizeInBytes_UnsupportedMapKey(t *testing.T) {
	_, err := GetInMemorySizeInBytes(map[interface{}]interface{}{make(chan int): "x"})
	if err == nil {
		t.Fatalf("expected error for unsupported map key")
	}
}

func TestGetInMemorySizeInBytes_PtrToUnsupported(t *testing.T) {
	ch := make(chan int)
	_, err := GetInMemorySizeInBytes(&ch)
	if err == nil {
		t.Fatalf("expected error for pointer to unsupported type")
	}
}

func TestGetPrimaryNetworkInterface_AddrsError(t *testing.T) {
	originalGetInterfaces := getInterfaces
	originalGetAddrs := getAddrs
	defer func() {
		getInterfaces = originalGetInterfaces
		getAddrs = originalGetAddrs
	}()

	mockInterface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp,
	}

	getInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{mockInterface}, nil
	}
	getAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return nil, fmt.Errorf("addr fail")
	}

	iface, addr, err := GetPrimaryNetworkInterface()
	if err == nil || iface != nil || addr != nil {
		t.Fatalf("expected error for addr failure")
	}
}

func TestGetPrimaryNetworkInterface_UnspecifiedIP(t *testing.T) {
	originalGetInterfaces := getInterfaces
	originalGetAddrs := getAddrs
	defer func() {
		getInterfaces = originalGetInterfaces
		getAddrs = originalGetAddrs
	}()

	mockInterface := net.Interface{
		Index: 1,
		Name:  "eth0",
		Flags: net.FlagUp,
	}

	getInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{mockInterface}, nil
	}
	getAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{&net.IPNet{IP: net.IPv4zero}}, nil
	}

	iface, addr, err := GetPrimaryNetworkInterface()
	if err == nil || iface != nil || addr != nil {
		t.Fatalf("expected error for unspecified ip")
	}
}

func TestGetAddrs_Default(t *testing.T) {
	_, _ = getAddrs(net.Interface{})
}
