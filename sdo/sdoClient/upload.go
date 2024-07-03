package sdoClient

import (
	"encoding/binary"
	"github.com/FabianPetersen/can"
	"github.com/FabianPetersen/canopen"
	"github.com/FabianPetersen/canopen/sdo"
	"strconv"

	"bytes"
	"time"
)

// Upload represents an SDO upload process to read data from a CANopen
// device – upload because the receiving node uploads data to another node.
type Upload struct {
	ObjectIndex canopen.ObjectIndex

	RequestCobID  uint16
	ResponseCobID uint16
}

func (upload Upload) Do(bus *can.Bus) ([]byte, error) {
	// Do not allow multiple messages for the same device
	key := strconv.Itoa(int(upload.RequestCobID))
	canopen.Lock.Lock(key)
	defer canopen.Lock.Unlock(key)

	c := &canopen.Client{Bus: bus, Timeout: time.Second * 2}
	// Initiate
	frame := canopen.Frame{
		CobID: upload.RequestCobID,
		Data: []byte{
			byte(sdo.InitiateUploadRequest << 5),
			upload.ObjectIndex.Index.B0, upload.ObjectIndex.Index.B1,
			upload.ObjectIndex.SubIndex,
			0x0, 0x0, 0x0, 0x0,
		},
	}

	req := canopen.NewRequest(frame, uint32(upload.ResponseCobID))
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	frame = resp.Frame
	scs := sdo.ServerCommandSpecifier(frame.Data[0] >> 5)
	if scs != sdo.InitiateUploadResponse {
		if scs == sdo.AbortTransfer {
			return nil, canopen.TransferAbort{
				AbortCode: sdo.GetAbortCodeBytes(frame),
			}
		} else {
			return nil, canopen.UnexpectedSCSResponse{
				Expected: 2,
				Actual:   uint8(scs),
			}
		}
	}

	// Check if this is the correct response for the requested message
	if !upload.ObjectIndex.Compare(frame.ObjectIndex()) {
		return nil, canopen.TransferAbort{}
	}

	if sdo.HasBit(frame.Data[0], 1) { // e = 1?
		// number of segment bytes with no data
		var n uint8
		if sdo.HasBit(frame.Data[0], 0) { // s = 1?
			n = (frame.Data[0] >> 2) & 0x3
		}
		return frame.Data[4 : 8-n], nil
	}

	// Read segment data length
	var total uint32
	b := bytes.NewBuffer(frame.Data[4:8])
	if err := binary.Read(b, binary.LittleEndian, &total); err != nil {
		return nil, err
	}

	var i = 0
	var buf bytes.Buffer
	for {
		data := make([]byte, 8)

		// ccs = 3
		data[0] = byte(sdo.UploadSegmentRequest << 5)

		if i%2 == 1 {
			// t = 1
			data[0] = sdo.SetBit(data[0], 4)
		}

		i += 1

		frame = canopen.Frame{
			CobID: upload.RequestCobID,
			Data:  data,
		}

		req = canopen.NewRequest(frame, uint32(upload.ResponseCobID))
		resp, err = c.DoMinDuration(req, 2*time.Millisecond)
		if err != nil {
			return nil, err
		}

		if sdo.HasBit(frame.Data[0], 4) != sdo.HasBit(resp.Frame.Data[0], 4) {
			return nil, canopen.UnexpectedToggleBit{
				Expected: sdo.HasBit(frame.Data[0], 4),
				Actual:   sdo.HasBit(resp.Frame.Data[0], 4),
			}
		}

		n := (resp.Frame.Data[0] >> 1) & 0x7
		buf.Write(resp.Frame.Data[1 : 8-n])

		// Check if we have received too many bytes
		if buf.Len() > int(total) {
			return nil, canopen.UnexpectedResponseLength{
				Expected: int(total),
				Actual:   buf.Len(),
			}
		}

		if sdo.HasBit(resp.Frame.Data[0], 0) { // c = 1?
			// Check if we have received too few bytes
			if buf.Len() != int(total) {
				return nil, canopen.UnexpectedResponseLength{
					Expected: int(total),
					Actual:   buf.Len(),
				}
			}
			break
		}
	}

	return buf.Bytes(), nil
}
