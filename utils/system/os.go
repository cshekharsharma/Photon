package system

import (
	"fmt"
	"net"
	"reflect"
	"runtime"
	"unsafe"
)

// Wrapper on runtime.GOOS, that tells you current OS name
var GetOperatingSystem = func() string {
	return runtime.GOOS
}

// IsLinux returns true if the current operating system is Linux.
//
// It uses the GOOS environment value to detect the platform at runtime,
// allowing platform-specific code branching if needed.
func IsLinux() bool {
	return GetOperatingSystem() == "linux"
}

// IsDarwin returns true if the current operating system is macOS (Darwin).
//
// This is useful when differentiating logic between macOS and other platforms
// like Linux or Windows.
func IsDarwin() bool {
	return GetOperatingSystem() == "darwin"
}

// IsWindows returns true if the current operating system is Windows.
//
// It checks the runtime GOOS value and can be used to write platform-specific
// code when behavior needs to vary between Windows and UNIX-like systems.
func IsWindows() bool {
	return GetOperatingSystem() == "windows"
}

// GetInMemorySizeInBytes estimates the in-memory size of a Go value.
//
// It handles primitive types (int, bool, float), strings, slices, maps,
// structs, and pointers recursively. For complex types like slices and maps,
// the size calculation includes both metadata (headers) and element data.
//
// Returns:
//   - int64: Total estimated size in bytes.
//   - error: If the type is unsupported.
//
// Note: This is an approximate estimation and may differ from actual
// runtime memory usage, especially for types with internal optimizations.
func GetInMemorySizeInBytes(v interface{}) (int64, error) {
	val := reflect.ValueOf(v)
	var stringHeader int64 = 16
	var sliceHeader int64 = 24
	var mapHeader int64 = 48

	switch val.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return int64(val.Type().Size()), nil

	case reflect.String:
		return stringHeader + int64(len(val.String())), nil

	case reflect.Slice:
		totalSize := sliceHeader
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			elemSize, err := GetInMemorySizeInBytes(elem)
			if err != nil {
				return 0, err
			}
			totalSize += elemSize
		}
		return totalSize, nil

	case reflect.Map:
		totalSize := mapHeader // Base map header size
		keys := val.MapKeys()

		for _, key := range keys {
			keySize, err := GetInMemorySizeInBytes(key.Interface())
			if err != nil {
				return 0, err
			}

			value := val.MapIndex(key).Interface()
			valueSize, err := GetInMemorySizeInBytes(value)
			if err != nil {
				return 0, err
			}

			totalSize += keySize + valueSize
		}
		return totalSize, nil

	case reflect.Pointer:
		if val.IsNil() {
			return 0, nil // Size is 0 for nil pointers
		}
		ptrSize := int64(unsafe.Sizeof(v))

		elemSize, err := GetInMemorySizeInBytes(val.Elem().Interface())
		if err != nil {
			return 0, err
		}
		return ptrSize + elemSize, nil

	case reflect.Struct:
		var structSize int64
		structSize = int64(val.Type().Size())
		return structSize, nil

	default:
		return 0, fmt.Errorf("unsupported type: %s", val.Kind())
	}
}

// net.Interfaces wrapper
var getInterfaces = net.Interfaces

// iface.Addrs() wrapper
var getAddrs = func(iface net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}

// GetPrimaryNetworkInterface returns the first non-loopback, active network interface
// and one of its associated IP addresses.
//
// It scans all available network interfaces and selects the first one that is "up"
// (running) and not a loopback interface. From there, it picks a usable IPv4 or IPv6
// address that is not unspecified or loopback.
//
// Returns:
//   - *net.Interface: Pointer to the selected network interface.
//   - net.Addr: The corresponding IP address.
//   - error: If no suitable interface is found or an error occurs.
//
// This function is useful for determining the primary external-facing IP address
// of the machine without needing external internet connections.
func GetPrimaryNetworkInterface() (*net.Interface, net.Addr, error) {
	interfaces, err := getInterfaces()
	if err != nil {
		return nil, nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := getAddrs(iface)
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				if ip.IsLoopback() || ip.IsUnspecified() {
					continue
				}
				return &iface, addr, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("no suitable network interface found")
}
