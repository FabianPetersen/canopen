package canopen

import (
	"encoding/binary"
	"fmt"
	"github.com/FabianPetersen/can"
	"github.com/jpillora/maplock"
	"time"
)

var Lock = maplock.New()

const (
	SDO_ERR_TOGGLE_BIT         = 0x05030000
	SDO_ERR_TIMEOUT            = 0x05040000
	SDO_ERR_COMMAND            = 0x05040001
	SDO_ERR_BLOCK_SIZE         = 0x05040002
	SDO_ERR_BLOCK_SEQUENCE     = 0x05040003
	SDO_ERR_BLOCK_CRC          = 0x05040004
	SDO_ERR_MEMORY             = 0x05040005
	SDO_ERR_ACCESS_UNSUPPORTED = 0x06010000
	SDO_ERR_ACCESS_WO          = 0x06010001
	SDO_ERR_ACCESS_RO          = 0x06010002
	SDO_ERR_NO_OBJECT          = 0x06020000
	SDO_ERR_MAPPING_OBJECT     = 0x06040041
	SDO_ERR_MAPPING_LENGTH     = 0x06040042
	SDO_ERR_GENERAL_PARAMETER  = 0x06040043
	SDO_ERR_GENERAL_DEVICE     = 0x06040047
	SDO_ERR_HARDWARE           = 0x06060000
	SDO_ERR_DATATYPE           = 0x06070010
	SDO_ERR_DATATYPE_HIGH      = 0x06070012
	SDO_ERR_DATATYPE_LOW       = 0x06070013
	SDO_ERR_NO_SUB_INDEX       = 0x06090011
	SDO_ERR_VALUE_RANGE        = 0x06090030
	SDO_ERR_VALUE_HIGH         = 0x06090031
	SDO_ERR_VALUE_LOW          = 0x06090032
	SDO_ERR_VALUE_MIN_MAX      = 0x06090036
	SDO_ERR_SDO_CONNECTION     = 0x060A0023
	SDO_ERR_GENERAL            = 0x08000000
	SDO_ERR_DATA_STORE         = 0x08000020
	SDO_ERR_DATA_STORE_LOCAL   = 0x08000021
	SDO_ERR_DATA_STORE_STATE   = 0x08000022
	SDO_ERR_OBJECT_DICTIONARY  = 0x08000023
	SDO_ERR_NO_DATA            = 0x08000024
)

func GetAbortCodeText(code uint32) string {
	switch code {
	case SDO_ERR_TOGGLE_BIT:
		return "SDO toggle bit error (protocol violation)"
	case SDO_ERR_TIMEOUT:
		return "SDO protocol timed out"
	case SDO_ERR_COMMAND:
		return "client/server command specifier not valid or unknown (protocol incompatibility)"
	case SDO_ERR_BLOCK_SIZE:
		return "Invalid block size (block mode only)"
	case SDO_ERR_BLOCK_SEQUENCE:
		return "Invalid sequence number (block mode only)"
	case SDO_ERR_BLOCK_CRC:
		return "CRC error (cyclic redundancy code, block mode only)"
	case SDO_ERR_MEMORY:
		return "out of memory"
	case SDO_ERR_ACCESS_UNSUPPORTED:
		return "unsupported access"
	case SDO_ERR_ACCESS_WO:
		return "tried to read a WRITE-ONLY object"
	case SDO_ERR_ACCESS_RO:
		return "tried to write a READ-ONLY object"
	case SDO_ERR_NO_OBJECT:
		return "object does not exist (in the CANopen object dictionary)"
	case SDO_ERR_MAPPING_OBJECT:
		return "object cannot be mapped (into a PDO)"
	case SDO_ERR_MAPPING_LENGTH:
		return "PDO length exceeded (when trying to map an object)"
	case SDO_ERR_GENERAL_PARAMETER:
		return "general parameter incompatibility"
	case SDO_ERR_GENERAL_DEVICE:
		return "general internal incompatibility in the device."
	case SDO_ERR_HARDWARE:
		return "access failed due to hardware error"
	case SDO_ERR_DATATYPE:
		return "data type and length code do not match"
	case SDO_ERR_DATATYPE_HIGH:
		return "data type problem, length code is too high"
	case SDO_ERR_DATATYPE_LOW:
		return "data type problem, length code is too low"
	case SDO_ERR_NO_SUB_INDEX:
		return "subindex does not exist"
	case SDO_ERR_VALUE_RANGE:
		return "value range exceeded"
	case SDO_ERR_VALUE_HIGH:
		return "value range exceeded, too high"
	case SDO_ERR_VALUE_LOW:
		return "value range exceeded, too low"
	case SDO_ERR_VALUE_MIN_MAX:
		return "maximum value is less than minimum value"
	case SDO_ERR_SDO_CONNECTION:
		return "resource not available: SDO connection"
	case SDO_ERR_GENERAL:
		return "general error"
	case SDO_ERR_DATA_STORE:
		return "data could not be transferred or stored"
	case SDO_ERR_DATA_STORE_LOCAL:
		return "data could not be transferred due to \"local control\""
	case SDO_ERR_DATA_STORE_STATE:
		return "data could not be transferred due to \"device state\""
	case SDO_ERR_OBJECT_DICTIONARY:
		return "object dictionary does not exist"
	case SDO_ERR_NO_DATA:
		return "no data"
	}
	return "unknown error"
}

type TransferAbort struct {
	AbortCode []uint8
}

func (e TransferAbort) Error() string {
	if len(e.AbortCode) == 4 {
		code := binary.LittleEndian.Uint32(e.AbortCode)
		return GetAbortCodeText(code)
	}

	return fmt.Sprintf("Server aborted upload")
}

type UnexpectedSCSResponse struct {
	Expected  uint8
	Actual    uint8
	AbortCode []uint8
}

func (e UnexpectedSCSResponse) Error() string {
	return fmt.Sprintf("unexpected server command specifier %X (expected %X)", e.Actual, e.Expected)
}

type UnexpectedResponseLength struct {
	Expected  int
	Actual    int
	AbortCode []uint8
}

func (e UnexpectedResponseLength) Error() string {
	return fmt.Sprintf("unexpected response length %X (expected %X)", e.Actual, e.Expected)
}

type UnexpectedToggleBit struct {
	Expected  bool
	Actual    bool
	AbortCode []uint8
}

func (e UnexpectedToggleBit) Error() string {
	return fmt.Sprintf("unexpected toggle bit %t (expected %t)", e.Actual, e.Expected)
}

// A Client handles message communication by sending a request
// and waiting for the response.
type Client struct {
	Bus     *can.Bus
	Timeout time.Duration
}

// Do sends a request and waits for a response.
// If the response frame doesn't arrive on time, an error is returned.
func (c *Client) Do(req *Request) (*Response, error) {
	return c.DoMinDuration(req, 10*time.Millisecond)
}

// DoMinDuration sends a request and waits for a response.
// If the response frame doesn't arrive on time, an error is returned.
func (c *Client) DoMinDuration(req *Request, min time.Duration) (*Response, error) {
	rch := can.Wait(c.Bus, req.ResponseID, c.Timeout)

	if err := c.Bus.PublishMinDuration(req.Frame.CANFrame(), min); err != nil {
		return nil, err
	}

	resp := <-rch

	return &Response{CANopenFrame(resp.Frame), req}, resp.Err
}
