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
	"strconv"
	"sync"

	"github.com/sourcenetwork/defradb/datastore"
)

var TxnStore sync.Map

//export createTransaction
func createTransaction(cIsConcurrent C.int, cIsReadOnly C.int) *C.Result {
	concurrent := cIsConcurrent != 0
	readOnly := cIsReadOnly != 0
	ctx := context.Background()
	var tx datastore.Txn
	var err error

	// Create a Txn object based on parameters passed in
	if concurrent {
		tx, err = globalNode.DB.NewConcurrentTxn(ctx, readOnly)
	} else {
		tx, err = globalNode.DB.NewTxn(ctx, readOnly)
	}
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingTxn, err), "")
	}

	// Store the Txn in the store, and return the ID to the user
	TxnStore.Store(tx.ID(), tx)
	IDstring := strconv.FormatUint(tx.ID(), 10)
	retVal := fmt.Sprintf(`{"id": %s}`, IDstring)
	return returnC(0, "", retVal)
}

//export commitTransaction
func commitTransaction(cTxnID C.ulonglong) *C.Result {
	TxnIDu64 := uint64(cTxnID)
	ctx := context.Background()

	// Get the transaction with the associated ID from the store
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	txn := tx.(datastore.Txn)

	// Commit the transaction, and if that doesn't error, remove it from the store
	err := txn.Commit(ctx)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	TxnStore.Delete(TxnIDu64)
	return returnC(0, "", "")
}

//export discardTransaction
func discardTransaction(cTxnID C.ulonglong) *C.Result {
	TxnIDu64 := uint64(cTxnID)
	ctx := context.Background()

	// Get the transaction with the associated ID from the store
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	txn := tx.(datastore.Txn)

	// Discard it, which currently cannot error, and then delete it from the store
	txn.Discard(ctx)
	TxnStore.Delete(TxnIDu64)
	return returnC(0, "", "")
}
