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

//export addSchema
func addSchema(cSchema *C.char) *C.Result {
	newSchema := C.GoString(cSchema)
	ctx := context.Background()
	_, err := globalNode.DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrAddingSchema, err), "")
	}
	return returnC(0, "", "")
}

//export describeSchema
func describeSchema(cName *C.char, cRoot *C.char, cVersion *C.char) *C.Result {
	ctx := context.Background()
	options := client.SchemaFetchOptions{}

	// Set the configuration options from the passed in parameters
	if cVersion != nil && C.GoString(cVersion) != "" {
		options.ID = immutable.Some(C.GoString(cVersion))
	}
	if cRoot != nil && C.GoString(cRoot) != "" {
		options.Root = immutable.Some(C.GoString(cRoot))
	}
	if cName != nil && C.GoString(cName) != "" {
		options.Name = immutable.Some(C.GoString(cName))
	}

	// Get the schema, and try to convert it to JSON for return
	schemas, err := globalNode.DB.GetSchemas(ctx, options)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingSchema, err), "")
	}
	bytes, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}

	jsonString := string(bytes)
	return returnC(0, "", jsonString)
}

//export patchSchema
func patchSchema(cPatch *C.char, cLensConfig *C.char, cSetActive C.int) *C.Result {
	ctx := context.Background()
	setActive := cSetActive != 0
	patchString, lensString := C.GoString(cPatch), C.GoString(cLensConfig)

	// Patch cannot be blank for this to work
	if cPatch == nil || patchString == "" {
		return returnC(1, cerrEmptyPatch, "")
	}

	// Set the lens configuration if it was passed in, and if it is valid
	decoder := json.NewDecoder(strings.NewReader(lensString))
	decoder.DisallowUnknownFields()
	var migration immutable.Option[model.Lens]
	if C.GoString(cLensConfig) != "" {
		var lensCfg model.Lens
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
		}
		migration = immutable.Some(lensCfg)
	}

	// Try to patch the schema, and return the result
	err := globalNode.DB.PatchSchema(ctx, patchString, migration, setActive)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrPatchingSchema, err), "")
	}
	return returnC(0, "", "")
}

//export setActiveSchema
func setActiveSchema(cVersion *C.char) *C.Result {
	ctx := context.Background()
	err := globalNode.DB.SetActiveSchemaVersion(ctx, C.GoString(cVersion))
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrSetActiveSchema, err), "")
	}
	return returnC(0, "", "")
}
