package mobilebindings

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
)

var globalNode *node.Node

func InitNode(dbPath string) int {
	ctx := context.Background()

	var err error

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

	storePath := filepath.Join("/storage/emulated/0/Android/data/com.example.gotestjava/files/", dbPath)
	dirPath := filepath.Dir(storePath)

	// Check if the directory exists; if not, create it
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755) // Create the directory with proper permissions
		if err != nil {
			println("Error creating store directory:", err.Error())
			return 1
		}
		println("Store directory created:", dirPath)
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

func StartNode() int {
	// Fail early if the node has not been initialized
	if globalNode == nil {
		println("Node is not initialized. Call initNode() first.")
		return 1
	}
	println("Here: a")
	ctx := context.Background()
	println("Here: b")
	err := globalNode.Start(ctx)
	println("Here: c")
	if err != nil {
		println("Error starting the node:", err.Error())
		return 1
	}
	println("Here: d")
	println("Node started successfully.")
	println("globalNode address: %p, type: %s\n", globalNode, reflect.TypeOf(globalNode))
	println("globalNode.DB address: %p, type: %s\n", globalNode.DB, reflect.TypeOf(globalNode.DB))
	return 0
}

func StopNode() int {
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

func AddSchema(newSchema string) int {
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

func AddDocument(colName string, jsonString string) int {
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

func ExecuteQuery(query string) string {

	println("Executing query:", query)

	if globalNode == nil {
		println("Node is not initialized. Call initNode() first.")
		return string("")
	}
	println("globalNode address: %p, type: %s\n", globalNode, reflect.TypeOf(globalNode))

	ctx := context.Background()

	if globalNode.DB == nil {
		println("Database is not initialized!")
		return string("")
	}
	println("globalNode.DB address: %p, type: %s\n", globalNode.DB, reflect.TypeOf(globalNode.DB))

	res := globalNode.DB.ExecRequest(ctx, query)

	// Check for errors in the GQL response, looping through and printing them all
	if len(res.GQL.Errors) > 0 {
		println("Error executing query:")
		for _, err := range res.GQL.Errors {
			println("Error:", err.Error())
		}
		return string("")
	}

	println("Query executed successfully.")

	// Try to marshall the JSON, and return it
	if res.GQL.Data != nil && len(res.GQL.Data.(map[string]any)) != 0 {
		dataJSON, err := json.Marshal(res.GQL.Data)
		if err != nil {
			println("Error marshalling Data to JSON:", err.Error())
			return string("")
		}
		return string(dataJSON)
	}
	return string("")
}

func GetString() string {
	return string("Hello")
}
