//go:build !windows

package netstat

func readInterfaceOctets() (inBytes, outBytes, wifiIn, wifiOut, otherIn, otherOut uint64, err error) {
	return 0, 0, 0, 0, 0, 0, nil
}
