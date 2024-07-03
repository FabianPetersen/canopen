package sdoServer

import (
	"encoding/binary"
	"github.com/FabianPetersen/canopen"
	"github.com/FabianPetersen/canopen/sdo"
)

func (server *Server) handleUpload(frame canopen.Frame) {
	// Read all the data
	objectIndex := frame.ObjectIndex()
	data, uploadErr := server.Upload(objectIndex)
	size := len(data)

	// Set scs = 2, s=1 (Always indicate size)
	headerByte := (byte(sdo.InitiateUploadResponse) << 5) + 1

	// Can be sent as expedited
	if uploadErr == canopen.NO_ERROR {
		if size <= 4 {
			server.expeditedUpload(headerByte, size, data, objectIndex)
		} else {
			server.ordinaryUpload(headerByte, size, data, objectIndex)
		}
	} else {
		server.publishError(uploadErr, objectIndex)
	}
}

func (server *Server) expeditedUpload(headerByte byte, size int, data []byte, objectIndex canopen.ObjectIndex) {
	// Set expedited bit
	headerByte += 2

	// Set empty bytes
	n := byte(4 - size)
	headerByte += n << 2

	// Pad the data always be 4 in length
	data = sdo.Pad(data, 4)

	server.publish(append(append([]byte{headerByte}, objectIndex.Bytes()...), data...))
}

func (server *Server) ordinaryUpload(headerByte byte, size int, data []byte, objectIndex canopen.ObjectIndex) {
	// Write the first frame
	sizeData := binary.LittleEndian.AppendUint32([]byte{}, uint32(size))
	currentFrameData := append(append([]byte{sdo.ServerResponseByte(sdo.InitiateUploadResponse, false)}, objectIndex.Bytes()...), sizeData...)

	// Go through all segments
	segments := sdo.SplitN(data, 7)
	for i, segmentData := range segments {
		// Wait for the client to respond
		resp, err := server.publishAndWait(currentFrameData)
		if err != nil {
			// Abort request (client timeout)
			server.publishError(canopen.SDO_ERR_TIMEOUT, objectIndex)
			return
		}

		ccs := sdo.ClientCommandSpecifier((resp.Frame.Data[0] >> 5) & 3)
		if ccs != sdo.UploadSegmentRequest {
			// Abort request (wrong ccs)
			server.publishError(canopen.SDO_ERR_COMMAND, objectIndex)
			return
		}

		// Send the next segment
		toggleBit := sdo.HasBit(resp.Frame.Data[0], 4)
		headerByte = sdo.ServerResponseByte(sdo.UploadSegmentResponse, toggleBit)

		// Set n (number of empty bytes)
		n := byte(7 - len(segmentData))
		headerByte += n << 1

		// Set c (1 == no more segments)
		if i == len(segments)-1 {
			headerByte += 1
		}
		// Pad the data to always have 7 bytes
		segmentData = sdo.Pad(segmentData, 7)

		// Prepare the next request
		currentFrameData = append([]byte{headerByte}, segmentData...)
	}

	// Send the last segment
	server.publish(currentFrameData)
}
