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
)

//export executeQuery
func executeQuery(cQuery *C.char) *C.Result {
	query := C.GoString(cQuery)
	ctx := context.Background()

	res := globalNode.DB.ExecRequest(ctx, query)

	// Caheck for errors in the GQL response, wrangling them into a single string
	if len(res.GQL.Errors) > 0 {
		var sb strings.Builder
		sb.WriteString("Error executing query:\n")
		for _, err := range res.GQL.Errors {
			sb.WriteString("Error: ")
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}
		return returnC(1, sb.String(), "")
	}

	// Try to marshall the JSON and return it
	dataMap, ok := res.GQL.Data.(map[string]any)
	if !ok || dataMap == nil || len(dataMap) == 0 {
		return returnC(1, "GraphQL response data is nil or empty.", "")
	}
	dataJSON, err := json.Marshal(dataMap)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(dataJSON))
}

func main() {}
