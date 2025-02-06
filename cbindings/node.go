//go:build cshared

package main

import (
	"C"
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
)

var globalNode *node.Node

//export initNode
func initNode(cPath *C.char) C.int {

	dbPath := C.GoString(cPath)
	ctx := context.Background()

	// If a database node already exists, close it first
	if globalNode != nil {
		println("Resetting existing node...")
		err := globalNode.Close(ctx)
		if err != nil {
			println("Error closing the node:", err)
			return 1
		}
		globalNode = nil
		println("Existing node successfully closed and reset.")
	}

	// Get a directory that is writable (i.e. in the Home Directory)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		println("Error getting home directory:", err.Error())
		return 1
	}
	storePath := filepath.Join(homeDir, dbPath)

	// Check if the directory exists, if not, create it
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		err := os.MkdirAll(storePath, 0755) // Create the directory with proper permissions
		if err != nil {
			println("Error creating store directory:", err.Error())
			return 1
		}
		println("Store directory created:", storePath)
	}

	// Create the database node with appropriate options
	globalNode, err = node.New(ctx, node.WithDisableP2P(true), node.WithDisableAPI(true), node.WithStorePath(storePath))
	if err != nil {
		println("Error creating the node:", err.Error())
		return 1
	}
	println("Node initialized.")
	return 0
}

//export startNode
func startNode() C.int {
	// Fail early if the node has not been initialized
	if globalNode == nil {
		println("Node is not initialized. Call initNode() first.")
		return 1
	}
	ctx := context.Background()
	err := globalNode.Start(ctx)
	if err != nil {
		println("Error starting the node:", err.Error())
		return 1
	}
	println("Node started successfully.")
	return 0
}

//export stopNode
func stopNode() C.int {
	if globalNode == nil {
		println("Node is not initialized or already stopped.")
		return 1
	}
	ctx := context.Background()
	err := globalNode.Close(ctx)
	if err != nil {
		println("Error stopping the node:", err)
		return 1
	}
	globalNode = nil
	println("Node stopped successfully.")
	return 0
}

//export addSchema
func addSchema(cSchema *C.char) C.int {
	newSchema := C.GoString(cSchema)
	ctx := context.Background()
	_, err := globalNode.DB.AddSchema(ctx, newSchema)
	if err != nil {
		println("Error:")
		println(err.Error())
		return 1
	}
	println("Added schema.")
	return 0
}

//export addDocument
func addDocument(cCollection *C.char, cJSON *C.char) C.int {
	colName := C.GoString(cCollection)
	jsonString := C.GoString(cJSON)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		println("Error:")
		println(err.Error())
		return 1
	}

	// If the collection exists, then try to create a new document
	// If this doesn't work, then fail out
	doc, err := client.NewDocFromJSON([]byte(jsonString), col.Definition())
	if err != nil {
		println("Error creating document:")
		println(err.Error())
		return 1
	}

	println("Document created")

	// Now we will try to insert the document into the collection
	err = col.Create(ctx, doc)
	if err != nil {
		println("Error inserting document:")
		println(err.Error())
		return 1
	}
	println("Document inserted successfully.")
	return 0
}

//export deleteDocument
func deleteDocument(cCollection *C.char, cDocID *C.char) C.int {
	colName := C.GoString(cCollection)
	docID := C.GoString(cDocID)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		println("Error:")
		println(err.Error())
		return 1
	}

	// Try to get a document ID from the string passed in
	docIDStruct, err := client.NewDocIDFromString(docID)
	if err != nil {
		println("Error creating document ID from string:")
		println(err.Error())
		return 1
	}

	// If the collection exists, then try to delete the document from it
	success, err := col.Delete(ctx, docIDStruct)
	if success != true {
		println("Error:")
		println(err.Error())
		return 1
	}

	println("Document deleted.")
	return 0
}

//export executeQuery
func executeQuery(cQuery *C.char) *C.char {
	query := C.GoString(cQuery)
	ctx := context.Background()

	res := globalNode.DB.ExecRequest(ctx, query)

	// Check for errors in the GQL response, looping through and printing them all
	if len(res.GQL.Errors) > 0 {
		println("Error executing query:")
		for _, err := range res.GQL.Errors {
			println("Error:", err.Error())
		}
		return C.CString("")
	}

	println("Query executed successfully.")

	// Try to marshall the JSON, and return it
	if res.GQL.Data != nil && len(res.GQL.Data.(map[string]any)) != 0 {
		dataJSON, err := json.Marshal(res.GQL.Data)
		if err != nil {
			println("Error marshalling Data to JSON:", err.Error())
			return C.CString("")
		}
		return C.CString(string(dataJSON))
	}
	return C.CString("")
}

func main() {}
