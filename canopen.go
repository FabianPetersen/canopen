package canopen

const (
	MessageTypeNMT       uint16 = 0x000
	MessageTypeSync      uint16 = 0x080
	MessageTypeTimestamp uint16 = 0x100
	MessageTypeTPDO1     uint16 = 0x180
	MessageTypeRPDO1     uint16 = 0x200
	MessageTypeTPDO2     uint16 = 0x280
	MessageTypeRPDO2     uint16 = 0x300
	MessageTypeTPDO3     uint16 = 0x380
	MessageTypeRPDO3     uint16 = 0x400
	MessageTypeTPDO4     uint16 = 0x480
	MessageTypeRPDO4     uint16 = 0x500
	// MessageTypeTSDO represents the type of SDO server response messages
	MessageTypeTSDO uint16 = 0x580
	// MessageTypeRSDO represents the type of SDO client request messages
	MessageTypeRSDO      uint16 = 0x600
	MessageTypeHeartbeat uint16 = 0x700
)

// MaxNodeID defines the highest node id
const MaxNodeID uint8 = 0x7F
const MPDO uint8 = 0x80

const (
	// MaskCobID is used to get 11 bits from an uint16 for the COB-ID
	MaskCobID = 0x7FF
	// MaskNodeID is used to extract the 7-bit node id from the COB-ID
	MaskNodeID = 0x7F
	// MaskMessageType is used to extract the 4-bit message type from the COB-ID
	MaskMessageType = 0x780

	// MaskIDSff is used to extract the valid 11-bit CAN identifier bits from the frame ID of a standard frame format.
	MaskIDSff = 0x000007FF
	// MaskIDEff is used to extract the valid 29-bit CAN identifier bits from the frame ID of an extended frame format.
	MaskIDEff = 0x1FFFFFFF
	// MaskErr is used to extract the the error flag (0 = data frame, 1 = error message) from the frame ID.
	MaskErr = 0x20000000
	// MaskRtr is used to extract the rtr flag (1 = rtr frame) from the frame ID
	MaskRtr = 0x40000000
	// MaskEff is used to extract the eff flag (0 = standard frame, 1 = extended frame) from the frame ID
	MaskEff = 0x80000000
)
