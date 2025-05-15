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
	"strconv"
	"time"

	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/node"
)

var globalNode *node.Node

//export nodeInit
func nodeInit(cOptions C.NodeInitOptions) *C.Result {
	dbPath := C.GoString(cOptions.dbPath)
	listeningAddresses := splitCommaSeparatedCString(cOptions.listeningAddresses)
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

	// Try to create the node options
	opts := []node.Option{
		node.WithStorePath(dbPath),
		node.WithLensRuntime(node.Wazero),
	}
	if len(listeningAddresses) > 0 {
		opts = append(opts, net.WithListenAddresses(listeningAddresses...))
	}
	maxTxnRetries := int(cOptions.maxTransactionRetries)
	if maxTxnRetries > 0 {
		opts = append(opts, db.WithMaxRetries(maxTxnRetries))
	}
	disableP2PFlag := cOptions.disableP2P != 0
	if disableP2PFlag {
		opts = append(opts, node.WithDisableP2P(true))
	}
	disableAPIFlag := cOptions.disableAPI != 0
	if disableAPIFlag {
		opts = append(opts, node.WithDisableAPI(true))
	}

	peers := splitCommaSeparatedCString(cOptions.peers)
	if len(peers) > 0 {
		opts = append(opts, net.WithBootstrapPeers(peers...))
	}
	// Configure the replicator retry times. Go from string slice -> time.Duration slice
	replicatorRetryTimes := splitCommaSeparatedCString(cOptions.replicatorRetryIntervals)
	var replicatorRetryIntervals []time.Duration
	for _, s := range replicatorRetryTimes {
		n, err := strconv.Atoi(s)
		if err != nil {
			return returnC(1, fmt.Sprintf(cerrParsingReplicatorTimes, err), "")
		}
		if n <= 0 {
			return returnC(1, cerrNegativeReplicatorTime, "")
		}
		replicatorRetryIntervals = append(replicatorRetryIntervals, time.Duration(n)*time.Second)
	}
	if len(replicatorRetryIntervals) > 0 {
		opts = append(opts, db.WithRetryInterval(replicatorRetryIntervals))
	}

	// Try to create the node passing in the collected options, then return the result
	globalNode, err = node.New(ctx, opts...)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingNode, err), "")
	}
	return returnC(0, "", "")
}

//export nodeStart
func nodeStart() *C.Result {
	if globalNode == nil {
		return returnC(1, cerrUninitializedNode, "")
	}
	ctx := context.Background()

	errCh := make(chan error, 1)

	go func() {
		err := globalNode.Start(ctx)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return returnC(1, err.Error(), "")
		}
		return returnC(0, "", "")
	case <-time.After(5 * time.Second):
		// Timeout occurred, node may still start later
		return returnC(2, "Node is still starting (timeout waiting for readiness)", "")
	}
}

//export nodeStop
func nodeStop() *C.Result {
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
