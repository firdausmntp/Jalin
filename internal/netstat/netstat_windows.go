//go:build windows

package netstat

import (
	"syscall"
	"unsafe"
)

var (
	modiphlpapi      = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfTable2  = modiphlpapi.NewProc("GetIfTable2")
	procFreeMibTable = modiphlpapi.NewProc("FreeMibTable")
)

// Layout must match MIB_IF_ROW2 byte for byte: Alias and Description are
// WCHAR[IF_MAX_STRING_SIZE+1] = 257, not 256. Shifting them by one WCHAR moves
// every counter and the status flags, which silently read as garbage.
type mibIfRow2 struct {
	InterfaceLuid               uint64
	InterfaceIndex              uint32
	InterfaceGuid               [16]byte
	Alias                       [257]uint16
	Description                 [257]uint16
	PhysicalAddressLength       uint32
	PhysicalAddress             [32]byte
	PermanentPhysicalAddress    [32]byte
	Mtu                         uint32
	Type                        uint32
	TunnelType                  uint32
	MediaType                   uint32
	PhysicalMediumType          uint32
	AccessType                  uint32
	DirectionType               uint32
	InterfaceAndOperStatusFlags uint32
	OperStatus                  uint32
	AdminStatus                 uint32
	MediaConnectState           uint32
	NetworkGuid                 [16]byte
	ConnectionType              uint32
	TransmitLinkSpeed           uint64
	ReceiveLinkSpeed            uint64
	InOctets                    uint64
	InUcastPkts                 uint64
	InNUcastPkts                uint64
	InDiscards                  uint64
	InErrors                    uint64
	InUnknownProtos             uint64
	InUcastOctets               uint64
	InMulticastOctets           uint64
	InBroadcastOctets           uint64
	OutOctets                   uint64
	OutUcastPkts                uint64
	OutNUcastPkts               uint64
	OutDiscards                 uint64
	OutErrors                   uint64
	OutUcastOctets              uint64
	OutMulticastOctets          uint64
	OutBroadcastOctets          uint64
	OutQLen                     uint64
}

type mibIfTable2 struct {
	NumEntries uint32
	Table      [1]mibIfRow2
}

func readInterfaceOctets() (inBytes, outBytes, wifiIn, wifiOut, otherIn, otherOut uint64, err error) {
	var pTable *mibIfTable2
	r1, _, errSys := procGetIfTable2.Call(uintptr(unsafe.Pointer(&pTable)))
	if r1 != 0 {
		return 0, 0, 0, 0, 0, 0, errSys
	}
	defer procFreeMibTable.Call(uintptr(unsafe.Pointer(pTable)))

	basePtr := unsafe.Pointer(&pTable.Table[0])
	rowSize := unsafe.Sizeof(mibIfRow2{})

	for i := uint32(0); i < pTable.NumEntries; i++ {
		row := (*mibIfRow2)(unsafe.Add(basePtr, uintptr(i)*rowSize))
		// Windows lists filter drivers (WFP, Npcap, VPN shims) as adapters that
		// mirror a real adapter's counters, so summing them all counts the same
		// bytes up to eight times. FilterInterface is the bit that marks them.
		const filterInterface = 0x02
		if row.InterfaceAndOperStatusFlags&filterInterface != 0 {
			continue
		}
		// Ignore software loopback (Type 24) and down interfaces.
		if row.Type != 24 && row.OperStatus == 1 {
			inBytes += row.InOctets
			outBytes += row.OutOctets
			if row.Type == 71 { // IF_TYPE_IEEE80211 (WiFi wireless)
				wifiIn += row.InOctets
				wifiOut += row.OutOctets
			} else { // Ethernet, cellular, or other physical/virtual adapters
				otherIn += row.InOctets
				otherOut += row.OutOctets
			}
		}
	}
	return inBytes, outBytes, wifiIn, wifiOut, otherIn, otherOut, nil
}
