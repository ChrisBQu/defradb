//go:build cgo
// +build cgo

package main

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/defradb/client"
)

//export P2PInfo
func P2PInfo() *C.Result {
	info := globalNode.DB.PeerInfo()
	return marshalJSONToCResult(info)
}

//export P2PgetAllReplicators
func P2PgetAllReplicators() *C.Result {
	ctx := context.Background()
	reps, err := globalNode.DB.GetAllReplicators(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return marshalJSONToCResult(reps)
}

//export P2PsetReplicator
func P2PsetReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	peerStr := C.GoString(cPeer)
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Create the replicator parameters
	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnC(1, err.Error(), "")
	}
	rep := client.ReplicatorParams{
		Info:        info,
		Collections: colArgs,
	}

	// Set the replicator and return the result
	err = globalNode.DB.SetReplicator(ctx, rep)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	peerStr := C.GoString(cPeer)
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Create the replicator parameters
	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnC(1, err.Error(), "")
	}
	rep := client.ReplicatorParams{
		Info:        info,
		Collections: colArgs,
	}

	// Set the replicator and return the result
	err = globalNode.DB.DeleteReplicator(ctx, rep)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export P2PcollectionAdd
func P2PcollectionAdd(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to add the collections, then return the result
	err = globalNode.DB.AddP2PCollections(ctx, colArgs)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export P2PcollectionRemove
func P2PcollectionRemove(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to remove the collections, then return the result
	err = globalNode.DB.RemoveP2PCollections(ctx, colArgs)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export P2PcollectionGetAll
func P2PcollectionGetAll(cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to get the collections, then return
	cols, err := globalNode.DB.GetAllP2PCollections(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return marshalJSONToCResult(cols)
}
