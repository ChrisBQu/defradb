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
	"github.com/sourcenetwork/immutable/enumerable"
)

//export lensSet
func lensSet(cSrc *C.char, cDst *C.char, cCfg *C.char) *C.Result {
	ctx := context.Background()
	srcSchemaVersionID := C.GoString(cSrc)
	dstSchemaVersionID := C.GoString(cDst)
	lensCfgJson := C.GoString(cCfg)

	// Parse the lens config string into a client.LensConfig
	decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
	}
	migrationCfg := client.LensConfig{
		SourceSchemaVersionID:      srcSchemaVersionID,
		DestinationSchemaVersionID: dstSchemaVersionID,
		Lens:                       lensCfg,
	}
	err := globalNode.DB.SetMigration(ctx, migrationCfg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export lensDown
func lensDown(cCollectionID C.uint, cDocuments *C.char) *C.Result {
	ctx := context.Background()
	collectionID := uint32(cCollectionID)
	srcData := []byte(C.GoString(cDocuments))

	// Decode the input documents
	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Call the lens down migration
	out, err := globalNode.DB.LensRegistry().MigrateDown(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Each reversed document will be appended to a value array
	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Marshall the value array and return it
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", string(jsonBytes))
}

//export lensUp
func lensUp(cCollectionID C.uint, cDocuments *C.char) *C.Result {
	ctx := context.Background()
	collectionID := uint32(cCollectionID)
	srcData := []byte(C.GoString(cDocuments))

	// Decode the input documents
	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Call the lens down migration
	out, err := globalNode.DB.LensRegistry().MigrateUp(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Each reversed document will be appended to a value array
	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Marshall the value array and return it
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", string(jsonBytes))
}

//export lensReload
func lensReload() *C.Result {
	ctx := context.Background()
	err := globalNode.DB.LensRegistry().ReloadLenses(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

//export lensSetRegistry
func lensSetRegistry(cCollectionID C.uint, cLensCfg *C.char) *C.Result {
	ctx := context.Background()
	collectionID := uint32(cCollectionID)
	lensCfgJSON := C.GoString(cLensCfg)

	// Create a model.Lens from the lens configuration
	decoder := json.NewDecoder(strings.NewReader(lensCfgJSON))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
	}

	// Set migration and return
	err := globalNode.DB.LensRegistry().SetMigration(ctx, collectionID, lensCfg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")

}
