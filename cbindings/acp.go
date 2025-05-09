//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"context"
)

//export acpAddPolicy
func acpAddPolicy(cIdentity *C.char, cPolicy *C.char) *C.Result {
	ctx := context.Background()
	policy := C.GoString(cPolicy)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to add the policy
	policyResult, err := globalNode.DB.AddPolicy(ctx, policy)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(policyResult)
}

//export acpAddRelationship
func acpAddRelationship(cIdentity *C.char, cCollection *C.char, cDocID *C.char, cRelation *C.char, cActor *C.char) *C.Result {
	ctx := context.Background()
	collectionArg := C.GoString(cCollection)
	docIDArg := C.GoString(cDocID)
	relationArg := C.GoString(cRelation)
	targetActorArg := C.GoString(cActor)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Add the relationship
	result, err := globalNode.DB.AddDocActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(result)
}

//export acpDeleteRelationship
func acpDeleteRelationship(cIdentity *C.char, cCollection *C.char, cDocID *C.char, cRelation *C.char, cActor *C.char) *C.Result {
	ctx := context.Background()
	collectionArg := C.GoString(cCollection)
	docIDArg := C.GoString(cDocID)
	relationArg := C.GoString(cRelation)
	targetActorArg := C.GoString(cActor)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Delete the relationship
	result, err := globalNode.DB.DeleteDocActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(result)
}
