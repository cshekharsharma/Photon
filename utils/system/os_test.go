package system

import (
	"fmt"
	"math"
	"net"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUintptrToInt64Overflow(t *testing.T) {
	if got := uintptrToInt64(uintptr(math.MaxUint)); got != math.MaxInt64 {
		t.Fatalf("expected MaxInt64 clamp, got %d", got)
	}
}

func TestGetSizeInBytes(t *testing.T) {
	type SampleStruct struct {
		A int
		B string
		C float64
	}

	testCases := []struct {
		name         string
		input        interface{}
		expectedSize int64
		expectError  bool
	}{
		{"TestInt", 42, 8, false},
		{"TestInt8", int8(42), 1, false},
		{"TestInt16", int16(42), 2, false},
		{"TestInt32", int32(42), 4, false},
		{"TestInt64", int64(42), 8, false},
		{"TestUint", uint(42), 8, false},
		{"TestUint8", uint8(42), 1, false},
		{"TestUint16", uint16(42), 2, false},
		{"TestUint32", uint32(42), 4, false},
		{"TestUint64", uint64(42), 8, false},
		{"TestFloat32", float32(3.14), 4, false},
		{"TestFloat64", float64(3.14), 8, false},
		{"TestBool", true, 1, false},

		{"TestString", "hello", int64(16 + len("hello")), false},
		{"TestEmpty", "", 16, false},

		{"TestIntSlice", []int{1, 2, 3, 4, 5}, 24 + 5*8, false},
		{"TestInt8Slice", []int8{1, 2, 3}, 24 + 3*1, false},
		{"TestBoolSlice", []bool{true, false, true}, 24 + 3, false},

		{"TestStringSlice", []string{"a", "bb", "hello world!"}, int64(24 + 16*3 + len("a") + len("bb") + len("hello world!")), false},
		{"TestEmptyStringSlice", []string{}, 24, false},

		{"TestIntSliceSlice", [][]int{{1, 2}, {3, 4}}, 24 + 24 + 2*8 + 24 + 2*8, false},
		{"TestEmptyIntSlice", []int{}, 24, false},
		{"TestSimpleMap", map[string]int{"a": 1, "b": 2}, 48 + (16 + 1) + 8 + (16 + 1) + 8, false}, // mapHeader + (key1 + value1) + (key2 + value2)

		{"TestStruct", SampleStruct{A: 1, B: "test", C: 3.14}, 32, false},
		{"TestPointerNonNil", func() *int { x := 10; return &x }(), 24, false},
		{"TestNilPointer", (*int)(nil), 0, false},
		{"TestUnsupportedType", make(chan int), 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualSize, err := GetInMemorySizeInBytes(tc.input)
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if actualSize != tc.expectedSize {
				t.Errorf("Expected %d bytes, but got %d bytes", tc.expectedSize, actualSize)
			}
		})
	}
}

func TestGetPrimaryNetworkInterface_Success(t *testing.T) {
	originalGetInterfaces := getInterfaces
	originalGetAddrs := getAddrs
	defer func() {
		getInterfaces = originalGetInterfaces
		getAddrs = originalGetAddrs
	}()

	mockInterface := net.Interface{
		Index: 1,
		MTU:   1500,
		Name:  "eth0",
		Flags: net.FlagUp,
	}

	mockAddr := &net.IPNet{IP: net.ParseIP("192.168.1.10"), Mask: net.CIDRMask(24, 32)}

	getInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{mockInterface}, nil
	}

	getAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{mockAddr}, nil
	}

	iface, addr, err := GetPrimaryNetworkInterface()

	assert.NoError(t, err)
	assert.Equal(t, "eth0", iface.Name)
	assert.Equal(t, mockAddr, addr)
}

func TestGetPrimaryNetworkInterface_NoInterfaces(t *testing.T) {
	originalGetInterfaces := getInterfaces
	defer func() { getInterfaces = originalGetInterfaces }()

	getInterfaces = func() ([]net.Interface, error) {
		return nil, fmt.Errorf("failed to get interfaces")
	}

	iface, addr, err := GetPrimaryNetworkInterface()

	assert.Error(t, err)
	assert.Nil(t, iface)
	assert.Nil(t, addr)
}

func TestGetPrimaryNetworkInterface_NoSuitableInterface(t *testing.T) {
	originalGetInterfaces := getInterfaces
	originalGetAddrs := getAddrs
	defer func() {
		getInterfaces = originalGetInterfaces
		getAddrs = originalGetAddrs
	}()

	getInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{
				Name:  "lo",
				Flags: net.FlagLoopback | net.FlagUp,
			},
		}, nil
	}

	getAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{}, nil
	}

	iface, addr, err := GetPrimaryNetworkInterface()

	assert.Error(t, err)
	assert.Nil(t, iface)
	assert.Nil(t, addr)
}
func TestGetOperatingSystem(t *testing.T) {
	assert.Equal(t, GetOperatingSystem(), runtime.GOOS)
}

func TestIsLinux(t *testing.T) {
	originalGetGOOS := GetOperatingSystem
	defer func() { GetOperatingSystem = originalGetGOOS }()

	GetOperatingSystem = func() string {
		return "linux"
	}

	assert.True(t, IsLinux())
	assert.False(t, IsDarwin())
}

func TestIsDarwin(t *testing.T) {
	originalGetGOOS := GetOperatingSystem
	defer func() { GetOperatingSystem = originalGetGOOS }()

	GetOperatingSystem = func() string {
		return "darwin"
	}

	assert.True(t, IsDarwin())
	assert.False(t, IsLinux())
}

func TestWindows(t *testing.T) {
	originalGetGOOS := GetOperatingSystem
	defer func() { GetOperatingSystem = originalGetGOOS }()

	GetOperatingSystem = func() string {
		return "windows"
	}

	assert.True(t, IsWindows())
	assert.False(t, IsLinux())
}

func TestIsNeither(t *testing.T) {
	originalGetGOOS := GetOperatingSystem
	defer func() { GetOperatingSystem = originalGetGOOS }()

	GetOperatingSystem = func() string {
		return "unknown"
	}

	assert.False(t, IsLinux())
	assert.False(t, IsDarwin())
	assert.False(t, IsWindows())
}
