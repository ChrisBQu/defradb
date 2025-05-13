//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/immutable"
)

type collectionContextKey struct{}
type schemaNameContextKey struct{}
type identityContextKey struct{}

// Helper function that attaches an identity to a context, returning the new context
func contextWithIdentity(ctx context.Context, privateKeyHex string) (context.Context, error) {
	if privateKeyHex == "" {
		return ctx, nil
	}
	data, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return ctx, err
	}
	privKey := secp256k1.PrivKeyFromBytes(data)
	newIdentity, err := identity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	if err != nil {
		return ctx, err
	}
	immutableIdentity := immutable.Some(newIdentity)
	newctx := identity.WithContext(ctx, immutableIdentity)
	return newctx, nil
}

// Helper function that attaches a transaction to a context, returning a new context
func contextWithTransaction(ctx context.Context, cTxnID C.ulonglong) (context.Context, error) {
	TxnIDu64 := uint64(cTxnID)
	if TxnIDu64 == 0 {
		return ctx, nil
	}
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return ctx, fmt.Errorf(cerrTxnDoesNotExist, TxnIDu64)
	}
	txn := tx.(datastore.Txn)
	ctx2 := context.WithValue(ctx, transactionContextKey{}, txn)
	return ctx2, nil
}

// Helper function that seeks to marshall JSON into a CResult
// The Result object will either contain the payload, if it works, or an error if it doesn't
func marshalJSONToCResult(value interface{}) *C.Result {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(dataJSON))
}
