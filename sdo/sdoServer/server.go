package sdoServer

import (
	"encoding/binary"
	"github.com/FabianPetersen/can"
	"github.com/FabianPetersen/canopen"
	"github.com/FabianPetersen/canopen/sdo"
	"time"
)

type Server struct {
	bus              *can.Bus
	client           *canopen.Client
	messageQueue     chan canopen.Frame
	clientRequestId  uint16
	serverResponseId uint16

	NodeId   uint8
	Upload   func(canopen.ObjectIndex) ([]byte, canopen.SDOAbortCode)
	Download func(canopen.ObjectIndex, []byte) canopen.SDOAbortCode
}

func (server *Server) Listen(bus *can.Bus) error {
	// Initial Setup
	server.bus = bus
	server.client = &canopen.Client{Bus: server.bus, Timeout: time.Second * 2}
	server.messageQueue = make(chan canopen.Frame, 500)
	server.clientRequestId = canopen.MessageTypeRSDO + uint16(server.NodeId)
	server.serverResponseId = canopen.MessageTypeTSDO + uint16(server.NodeId)

	// Setup function to listen to new requests
	server.setupListener()

	// Open a new thread to process messages
	go server.processMessageQueue()

	// Publish and subscribe
	return server.bus.ConnectAndPublish()
}

func (server *Server) setupListener() {
	// Setup listener
	server.bus.SubscribeFunc(func(frame can.Frame) {
		coFrame := canopen.CANopenFrame(frame)

		// Check if the frame is intended for us and is SDO
		if coFrame.NodeID() == server.NodeId && coFrame.MessageType() == canopen.MessageTypeRSDO && len(coFrame.Data) == 8 {
			// Check that it is a new SDO request
			ccs, _, _, _ := sdo.ProcessRequestByte(coFrame.Data[0])
			if ccs == sdo.InitiateUploadRequest || ccs == sdo.InitiateDownloadRequest {
				server.messageQueue <- coFrame
			}
		}
	})
}

func (server *Server) processMessageQueue() {
	for coFrame := range server.messageQueue {
		ccs, _, _, _ := sdo.ProcessRequestByte(coFrame.Data[0])

		// Is the frame upload or download
		if ccs == sdo.InitiateUploadRequest {
			server.handleUpload(coFrame)

		} else if ccs == sdo.InitiateDownloadRequest {
			server.handleDownload(coFrame)
		}
	}
}

func (server *Server) publishError(errorCode canopen.SDOAbortCode, objectIndex canopen.ObjectIndex) error {
	errorData := binary.LittleEndian.AppendUint32([]byte{}, uint32(errorCode))
	return server.publish(append(append([]byte{byte(sdo.AbortTransfer << 5)}, objectIndex.Bytes()...), errorData...))
}

func (server *Server) publish(payload []byte) error {
	// Pad the result to always have 8 bytes
	payload = sdo.Pad(payload, 8)

	return server.bus.Publish(can.Frame{
		ID:     uint32(server.serverResponseId),
		Length: 8,
		Data: [8]byte{
			payload[0], payload[1], payload[2], payload[3], payload[4], payload[5], payload[6], payload[7],
		},
	})
}

func (server *Server) publishAndWait(payload []byte) (*canopen.Response, error) {
	// Pad the result to always have 8 bytes
	payload = sdo.Pad(payload, 8)

	req := canopen.NewRequest(canopen.NewFrame(server.serverResponseId, payload), uint32(server.clientRequestId))
	return server.client.Do(req)
}
