package sdo

import "github.com/FabianPetersen/canopen"

func HasBit(n uint8, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func SetBit(n uint8, pos uint) uint8 {
	n |= (1 << pos)
	return n
}

func GetAbortCodeBytes(frame canopen.Frame) []uint8 {
	if len(frame.Data) >= 8 {
		return frame.Data[4:]
	}
	return []uint8{}
}

// SplitN splits b into a list of n sized bytes
func SplitN(b []byte, n int) [][]byte {
	if len(b) < n {
		return [][]byte{b}
	}

	var bs [][]byte
	var buf []byte
	for i := 0; i < len(b); i++ {
		if len(buf) == n {
			bs = append(bs, buf)
			buf = []byte{}
		}

		buf = append(buf, b[i])
	}

	if len(buf) > 0 {
		bs = append(bs, buf)
	}

	return bs
}

func Pad(b []byte, minLength int) []byte {
	for i := len(b); i < minLength; i++ {
		b = append(b, 0)
	}

	return b
}

type ClientCommandSpecifier byte

const (
	InitiateDownloadRequest ClientCommandSpecifier = 1
	DownloadSegmentRequest  ClientCommandSpecifier = 0
	ClientBlockDownload     ClientCommandSpecifier = 6

	InitiateUploadRequest ClientCommandSpecifier = 2
	UploadSegmentRequest  ClientCommandSpecifier = 3
)

type ServerCommandSpecifier byte

const (
	InitiateDownloadResponse ServerCommandSpecifier = 3
	DownloadSegmentResponse  ServerCommandSpecifier = 1
	ServerBlockDownload      ServerCommandSpecifier = 5

	InitiateUploadResponse ServerCommandSpecifier = 2
	UploadSegmentResponse  ServerCommandSpecifier = 0

	AbortTransfer ServerCommandSpecifier = 4
)

func ProcessRequestByte(clientData byte) (ClientCommandSpecifier, bool, bool, byte) {
	clientCommandSpecifier := clientData >> 5
	isExpedited := HasBit(clientData, 1)
	hasSize := HasBit(clientData, 0)
	n := (clientData >> 2) & 3
	return ClientCommandSpecifier(clientCommandSpecifier), isExpedited, hasSize, n
}

func ServerResponseByte(scs ServerCommandSpecifier, toggleBit bool) byte {
	toggleValue := 0
	if toggleBit {
		toggleValue = 1
	}

	return (byte(scs) << 5) + (byte(toggleValue) << 4)
}
