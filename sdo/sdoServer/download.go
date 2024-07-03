package sdoServer

import (
	"encoding/binary"
	"github.com/FabianPetersen/canopen"
	"github.com/FabianPetersen/canopen/sdo"
)

func (server *Server) handleDownload(frame canopen.Frame) {
	// Check if it is expedited download
	_, isExpedited, _, _ := sdo.ProcessRequestByte(frame.Data[0])

	// Check if the data is expedited (data in frame)
	if isExpedited {
		server.expeditedDownload(frame)
	} else {
		server.ordinaryDownload(frame)
	}
}

func (server *Server) expeditedDownload(frame canopen.Frame) {
	_, _, _, n := sdo.ProcessRequestByte(frame.Data[0])

	// The data as specified by n (number of bytes that does not contain data)
	data := frame.Data[4 : 8-n]

	// Process the data
	objectIndex := frame.ObjectIndex()
	downloadError := server.Download(objectIndex, data)

	// Accept the download request
	if downloadError == canopen.NO_ERROR {
		_ = server.publish([]byte{sdo.ServerResponseByte(sdo.InitiateDownloadResponse, false)})
	} else {
		server.publishError(downloadError, objectIndex)
	}
}

func (server *Server) ordinaryDownload(frame canopen.Frame) {
	_, _, _, n := sdo.ProcessRequestByte(frame.Data[0])
	objectIndex := frame.ObjectIndex()

	// Data 4-7 contains the number of bytes to be downloaded
	size := binary.LittleEndian.Uint32(frame.Data[4:])

	// Accept the request, to get the actual data from the client
	resp, err := server.publishAndWait([]byte{sdo.ServerResponseByte(sdo.InitiateDownloadResponse, false)})
	if err != nil {
		// The client did not respond in time
		// TODO abort the request
		return
	}

	completeData := []byte{}
	noContinueBit := false
	toggleBit := false
	lastToggleBit := true

	// Continue until the client indicates that there are not more parts
	for !noContinueBit {
		ccs := sdo.ClientCommandSpecifier((resp.Frame.Data[0] >> 5) & 3)
		if ccs == sdo.DownloadSegmentRequest {
			toggleBit = sdo.HasBit(resp.Frame.Data[0], 4)
			if toggleBit == lastToggleBit {
				// Abort the request (toggle bit not set correctly)
				server.publishError(canopen.SDO_ERR_TOGGLE_BIT, objectIndex)
				return
			}

			n = (resp.Frame.Data[0] >> 1) & 7 // Number of bytes in d that does not contain data
			noContinueBit = sdo.HasBit(resp.Frame.Data[0], 0)
			segData := resp.Frame.Data[1 : 8-n]

			// Append the data
			completeData = append(completeData, segData...)

			if !noContinueBit {
				// Request the next part
				resp, err = server.publishAndWait([]byte{sdo.ServerResponseByte(sdo.DownloadSegmentResponse, toggleBit)})
				if err != nil {
					// Abort the request (no response)
					server.publishError(canopen.SDO_ERR_TIMEOUT, objectIndex)
					return
				}
			}

			lastToggleBit = toggleBit
		} else {
			// Abort the request (ccs not correct)
			server.publishError(canopen.SDO_ERR_COMMAND, objectIndex)
			return
		}
	}

	// Process the data
	downloadError := server.Download(frame.ObjectIndex(), completeData[0:size])

	// Sent the response
	if downloadError == canopen.NO_ERROR {
		_ = server.publish([]byte{sdo.ServerResponseByte(sdo.DownloadSegmentResponse, toggleBit)})

	} else {
		server.publishError(downloadError, objectIndex)
	}
}
