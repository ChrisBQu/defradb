//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

//export identityNew
func identityNew(cKeyType *C.char) *C.Result {
	keyTypeStr := C.GoString(cKeyType)

	// Create a public key object of the specified type (Secp256k1 by default) and use it to create identity
	keyType := crypto.KeyTypeSecp256k1
	if keyTypeStr != "" {
		keyType = crypto.KeyType(keyTypeStr)
	}
	newIdentity, err := identity.Generate(crypto.KeyType(keyType))
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(newIdentity.IntoRawIdentity())
}
