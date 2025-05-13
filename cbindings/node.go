//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"context"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/node"
)

// Helper function which builds a return struct from Go to C
func returnC(status int, errortext string, valuetext string) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(status)
	result.error = C.CString(errortext)
	result.value = C.CString(valuetext)

	return result
}

var globalNode *node.Node
var nodeReady = make(chan struct{})

var listeningIPs = []string{"/ip4/127.0.0.1/tcp/9171"}

//export initNode
func initNode(cPath *C.char) *C.Result {
	dbPath := C.GoString(cPath)
	ctx := context.Background()

	if globalNode != nil {
		err := globalNode.Close(ctx)
		if err != nil {
			return returnC(1, fmt.Sprintf(cerrClosingNode, err), "")
		}
		globalNode = nil
	}

	// Create the directory if it doesn't exist
	var err error
	if _, err = os.Stat(dbPath); os.IsNotExist(err) {
		err := os.MkdirAll(dbPath, 0755)
		if err != nil {
			return returnC(1, fmt.Sprintf(cerrCreatingStoreDirectory, err), "")
		}
	}

	// Try to create the node
	globalNode, err = node.New(
		ctx,
		node.WithStorePath(dbPath),
		node.WithLensRuntime(node.Wazero),
		net.WithListenAddresses(listeningIPs...),
	)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingNode, err), "")
	}

	return returnC(0, "", "")
}

//export startNode
func startNode() *C.Result {

	// Fail early if the node has not been initialized
	if globalNode == nil {
		return returnC(1, cerrUninitializedNode, "")
	}
	ctx := context.Background()

	go func() {
		err := globalNode.Start(ctx)
		if err != nil {
			// TO DO: Pass this out to the main function for return to C
			println("Failed to start node")
			return
		}
		close(nodeReady)
	}()

	// Wait for the node to be ready, or timeout after 5 seconds
	select {
	case <-nodeReady:
		// Node started successfully
		return returnC(0, "", "")
	case <-time.After(5 * time.Second):
		// Timeout occurred
		return returnC(1, "Timed out waiting for node to start", "")
	}
}

//export stopNode
func stopNode() *C.Result {
	if globalNode == nil {
		return returnC(1, cerrStoppedNode, "")
	}
	ctx := context.Background()
	err := globalNode.Close(ctx)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrStoppingNode, err), "")
	}
	globalNode = nil

	return returnC(0, "", "")
}
