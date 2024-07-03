package sdo

import (
	"bytes"
	"fmt"
	"github.com/FabianPetersen/can"
	"github.com/FabianPetersen/canopen"
	"github.com/FabianPetersen/canopen/sdo/sdoClient"
	"github.com/FabianPetersen/canopen/sdo/sdoServer"
	"testing"
	"time"
)

func getBus() *can.Bus {
	canbus, _ := can.NewBusForInterfaceWithName(fmt.Sprintf("can%d", 0))

	go func() {
		_ = canbus.ConnectAndPublish()
	}()

	time.Sleep(100 * time.Millisecond)
	return canbus
}

var server sdoServer.Server

func Setup(t *testing.T, response chan []byte) error {
	t.Log("Start server")

	upload := func(index canopen.ObjectIndex) ([]byte, canopen.SDOAbortCode) {
		return <-response, canopen.NO_ERROR
	}

	download := func(index canopen.ObjectIndex, bytes []byte) canopen.SDOAbortCode {
		response <- bytes
		t.Log("Download data in server", bytes)
		return canopen.NO_ERROR
	}

	if server.NodeId == 0 {
		// Setup server
		server = sdoServer.Server{
			NodeId:   1,
			Upload:   upload,
			Download: download,
		}

		canbus, _ := can.NewBusForInterfaceWithName(fmt.Sprintf("can%d", 0))
		return server.Listen(canbus)
	} else {
		server.Upload = upload
		server.Download = download
		return nil
	}
}

func TestExpeditedDownload(t *testing.T) {
	response := make(chan []byte, 1)
	go func() {
		_ = Setup(t, response)
	}()

	// Send the request
	sentData := []byte{0, 1, 2, 3}
	clientErr := sdoClient.Download{
		ObjectIndex:   canopen.ObjectIndex{},
		Data:          sentData,
		RequestCobID:  canopen.MessageTypeRSDO + 1,
		ResponseCobID: canopen.MessageTypeTSDO + 1,
	}.Do(getBus())

	if clientErr != nil {
		t.Log("Client error", clientErr)
		t.FailNow()
	}

	receivedData := <-response
	if !bytes.Equal(receivedData, sentData) {
		t.Log("Data does not match", sentData, receivedData)
		t.FailNow()
	}
}

func TestDownload(t *testing.T) {
	response := make(chan []byte, 1)
	go func() {
		_ = Setup(t, response)
	}()

	// Send the request
	sentData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	clientErr := sdoClient.Download{
		ObjectIndex:   canopen.ObjectIndex{},
		Data:          sentData,
		RequestCobID:  canopen.MessageTypeRSDO + 1,
		ResponseCobID: canopen.MessageTypeTSDO + 1,
	}.Do(getBus())

	if clientErr != nil {
		t.Log("Client", clientErr)
		t.FailNow()
	}

	receivedData := <-response
	if !bytes.Equal(receivedData, sentData) {
		t.Log("Data does not match", sentData, receivedData)
		t.FailNow()
	}
}

func TestExpeditedUpload(t *testing.T) {
	response := make(chan []byte, 1)
	sentData := []byte{0, 1, 2, 3}
	response <- sentData
	go func() {
		_ = Setup(t, response)
	}()

	// Send the request
	receivedData, clientErr := sdoClient.Upload{
		ObjectIndex:   canopen.ObjectIndex{},
		RequestCobID:  canopen.MessageTypeRSDO + 1,
		ResponseCobID: canopen.MessageTypeTSDO + 1,
	}.Do(getBus())

	if clientErr != nil {
		t.Log("Client", clientErr)
		t.FailNow()
	}

	if !bytes.Equal(receivedData, sentData) {
		t.Log("Data does not match", sentData, receivedData)
		t.FailNow()
	}
}

func TestUpload(t *testing.T) {
	response := make(chan []byte, 1)
	sentData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	response <- sentData
	go func() {
		_ = Setup(t, response)
	}()

	// Send the request
	receivedData, clientErr := sdoClient.Upload{
		ObjectIndex:   canopen.ObjectIndex{},
		RequestCobID:  canopen.MessageTypeRSDO + 1,
		ResponseCobID: canopen.MessageTypeTSDO + 1,
	}.Do(getBus())

	if clientErr != nil {
		t.Log("Client", clientErr)
		t.FailNow()
	}

	if !bytes.Equal(receivedData, sentData) {
		t.Log("Data does not match", sentData, receivedData)
		t.FailNow()
	}
}
