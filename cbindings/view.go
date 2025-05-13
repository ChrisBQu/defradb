//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

//export viewAdd
func viewAdd(cQuery *C.char, cSDL *C.char, cTransform *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	query := C.GoString(cQuery)
	sdl := C.GoString(cSDL)
	lensCfgJson := C.GoString(cTransform)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Decode the lens configuration into a lens
	var transform immutable.Option[model.Lens]
	if lensCfgJson != "" {
		decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
		decoder.DisallowUnknownFields()
		var lensCfg model.Lens
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
		}
		transform = immutable.Some(lensCfg)
	}

	// Add the view
	defs, err := globalNode.DB.AddView(ctx, query, sdl, transform)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(defs)
}

//export viewRefresh
func viewRefresh(cViewName *C.char, cSchemaRoot *C.char, cVersionID *C.char, cGetInactive C.int, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	name := C.GoString(cViewName)
	schemaRoot := C.GoString(cSchemaRoot)
	versionID := C.GoString(cVersionID)
	getInactive := cGetInactive != 0

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Parse the input parameters into a CollectionFetchOptions object
	options := client.CollectionFetchOptions{}
	if versionID != "" {
		options.SchemaVersionID = immutable.Some(versionID)
	}
	if schemaRoot != "" {
		options.SchemaRoot = immutable.Some(schemaRoot)
	}
	if name != "" {
		options.Name = immutable.Some(name)
	}
	if getInactive {
		options.IncludeInactive = immutable.Some(getInactive)
	}

	// Refresh the views and return
	err = globalNode.DB.RefreshViews(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}
