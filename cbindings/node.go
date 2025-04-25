//go:build cgo
// +build cgo

package main

/*
typedef struct {
    int status;
    char* error;
	char* value;
} Result;
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
)

// Helper function which builds a return struct from Go to C
func returnC(status int, errortext string, valuetext string) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(status)
	result.error = C.CString(errortext)
	result.value = C.CString(valuetext)

	return result
}

var globalNode *node.Node

//export initNode
func initNode(cPath *C.char) *C.Result {
	dbPath := C.GoString(cPath)
	ctx := context.Background()

	if globalNode != nil {
		err := globalNode.Close(ctx)
		if err != nil {
			return returnC(1, fmt.Sprintf("Error closing the node: %v", err), "")
		}
		globalNode = nil
	}

	// Create the directory if it doesn't exist
	var err error
	if _, err = os.Stat(dbPath); os.IsNotExist(err) {
		err := os.MkdirAll(dbPath, 0755)
		if err != nil {
			return returnC(1, fmt.Sprintf("Error creating the store directory: %v", err), "")
		}
	}

	// Try to create the node
	globalNode, err = node.New(
		ctx,
		node.WithDisableP2P(true),
		node.WithDisableAPI(true),
		node.WithStorePath(dbPath),
		node.WithLensRuntime(node.Wazero),
	)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error creating the node: %v", err), "")
	}

	return returnC(0, "", "")
}

//export startNode
func startNode() *C.Result {

	// Fail early if the node has not been initialized
	if globalNode == nil {
		return returnC(1, "Node is not initialized. Call initNode() first.", "")
	}
	ctx := context.Background()
	err := globalNode.Start(ctx)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error starting the node: %v", err), "")
	}

	return returnC(0, "", "")
}

//export stopNode
func stopNode() *C.Result {
	if globalNode == nil {
		return returnC(1, "Node is not initialized or was already", "")
	}
	ctx := context.Background()
	err := globalNode.Close(ctx)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error stopping the node: %v", err), "")
	}
	globalNode = nil

	return returnC(0, "", "")
}

//export addSchema
func addSchema(cSchema *C.char) *C.Result {
	newSchema := C.GoString(cSchema)
	ctx := context.Background()
	_, err := globalNode.DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error: %v", err), "")
	}
	return returnC(0, "", "")
}

//export addDocument
func addDocument(cCollection *C.char, cJSON *C.char) *C.Result {
	colName := C.GoString(cCollection)
	jsonString := C.GoString(cJSON)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error: %v", err), "")
	}

	// If the collection exists, then try to create a new document
	// If this doesn't work, then fail out
	doc, err := client.NewDocFromJSON([]byte(jsonString), col.Definition())
	if err != nil {
		return returnC(1, fmt.Sprintf("Error creating document: %v", err), "")
	}

	// Now we will try to insert the document into the collection
	err = col.Create(ctx, doc)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error inserting document: %v", err), "")
	}
	return returnC(0, "", "")
}

//export deleteDocument
func deleteDocument(cCollection *C.char, cDocID *C.char) *C.Result {
	colName := C.GoString(cCollection)
	docID := C.GoString(cDocID)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error: %v", err), "")
	}

	// Try to get a document ID from the string passed in
	docIDStruct, err := client.NewDocIDFromString(docID)
	if err != nil {
		return returnC(1, fmt.Sprintf("Error creating document ID from string: %v", err), "")
	}

	// If the collection exists, then try to delete the document from it
	success, err := col.Delete(ctx, docIDStruct)
	if success != true {
		return returnC(1, fmt.Sprintf("Error: %v", err), "")
	}

	return returnC(0, "", "")
}

//export executeQuery
func executeQuery(cQuery *C.char) *C.Result {
	query := C.GoString(cQuery)
	ctx := context.Background()

	res := globalNode.DB.ExecRequest(ctx, query)

	// Check for errors in the GQL response, wrangling them into a single string
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

	// Try to marshall the JSON, and return it
	if res.GQL.Data != nil && len(res.GQL.Data.(map[string]any)) != 0 {
		dataJSON, err := json.Marshal(res.GQL.Data)
		if err != nil {
			return returnC(1, fmt.Sprintf("Error marshalling data to JSON: %v", err), "")
		}
		return returnC(0, "", string(dataJSON))
	}

	return returnC(0, "", "")
}

func main() {}
