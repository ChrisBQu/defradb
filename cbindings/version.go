//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"github.com/sourcenetwork/defradb/version"
)

//export versionGet
func versionGet(cFlagFull C.int, cFlagJSON C.int) *C.Result {
	flagFull := cFlagFull != 0
	flagJSON := cFlagJSON != 0

	// Call the version function
	dv, err := version.NewDefraVersion()
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Return either the JSON, the long string version, or the short string version
	if flagJSON {
		return marshalJSONToCResult(dv)
	}
	if flagFull {
		return returnC(0, "", dv.StringFull())
	}
	return returnC(0, "", dv.String())
}
