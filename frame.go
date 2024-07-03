package canopen

import (
	"github.com/FabianPetersen/can"
)

// A Frame represents a CANopen frame.
type Frame struct {
	// CobID is the 11-bit communication object identifier – CANopen only uses 11-bit identifiers.
	// Bits 0-6 represent the 7-bit node ID. Bits 7-11 represent the 4-bit message type.
	CobID uint16
	// Rtr represents the Remote Transmit Request flag.
	Rtr bool
	// Data contains 8 bytes
	Data []uint8
}

// CANopenFrame returns a CANopen frame from a CAN frame.
func CANopenFrame(frm can.Frame) Frame {
	canopenFrame := Frame{}

	canopenFrame.CobID = uint16(frm.ID & MaskIDSff)
	canopenFrame.Rtr = (frm.ID & MaskRtr) == MaskRtr
	canopenFrame.Data = frm.Data[:]

	return canopenFrame
}

// NewFrame returns a frame with an id and data bytes.
func NewFrame(id uint16, data []uint8) Frame {
	return Frame{
		CobID: id & MaskCobID, // only use first 11 bits
		Data:  data,
	}
}

// MessageType returns the message type.
func (frm Frame) MessageType() uint16 {
	return frm.CobID & MaskMessageType
}

// NodeID returns the node id.
func (frm Frame) NodeID() uint8 {
	return uint8(frm.CobID & MaskNodeID)
}

// CANFrame returns a CAN frame representing the CANopen frame.
//
// CANopen frames are encoded as follows:
//
//	         -------------------------------------------------------
//	CAN     | ID           | Length    | Flags | Res0 | Res1 | Data |
//	         -------------------------------------------------------
//	CANopen | COB-ID + Rtr | len(Data) |       |      |      | Data |
//	         -------------------------------------------------------
func (frm Frame) CANFrame() can.Frame {
	var data [8]uint8
	n := len(frm.Data)
	copy(data[:n], frm.Data[:n])

	// Convert CANopen COB-ID to CAN id including RTR flag
	id := uint32(frm.CobID)
	if frm.Rtr == true {
		id = id | MaskRtr
	}

	return can.Frame{
		ID:     id,
		Length: uint8(len(frm.Data)),
		Data:   data,
	}
}

func (frm *Frame) ObjectIndex() ObjectIndex {
	if len(frm.Data) < 4 {
		return ObjectIndex{}
	}

	return ObjectIndex{
		Index: Index{
			B0: frm.Data[1],
			B1: frm.Data[2],
		},
		SubIndex: frm.Data[3],
	}
}
