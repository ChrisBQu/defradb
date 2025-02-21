//go:build cshared

package main

/*
#cgo CFLAGS: -I${SRCDIR}/jni

#include <jni.h>
#include <stdlib.h>

static const char* GetStringUTFChars(JNIEnv *env, jstring str) {
    return (*env)->GetStringUTFChars(env, str, 0);
}

static void ReleaseStringUTFChars(JNIEnv *env, jstring str, const char* chars) {
    (*env)->ReleaseStringUTFChars(env, str, chars);
}

static jstring NewStringUTF(JNIEnv *env, const char* chars) {
    return (*env)->NewStringUTF(env, chars);
}
*/
import "C"
import (
	"context"
	"os"
	"path/filepath"
	"unsafe"

	"encoding/json"

	"github.com/sourcenetwork/defradb/node"

	"github.com/sourcenetwork/defradb/client"
)

var globalNode *node.Node

func makeJString(env *C.JNIEnv, str string) unsafe.Pointer {
	cstr := C.CString(str) // Allocate C string

	jstr := C.NewStringUTF(env, cstr) // Convert to jstring
	C.free(unsafe.Pointer(cstr))      // Free C string after use

	return unsafe.Pointer(jstr)
}

//export Java_com_example_gotestjava_MainActivity_initNode
func Java_com_example_gotestjava_MainActivity_initNode(env unsafe.Pointer, obj unsafe.Pointer, input unsafe.Pointer) C.int {
	cstr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(input))
	dbPath := C.GoString(cstr)
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

//export Java_com_example_gotestjava_MainActivity_StartNode
func Java_com_example_gotestjava_MainActivity_StartNode(env unsafe.Pointer, obj unsafe.Pointer) C.int {
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

//export Java_com_example_gotestjava_MainActivity_StopNode
func Java_com_example_gotestjava_MainActivity_StopNode(env unsafe.Pointer, obj unsafe.Pointer) C.int {
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

//export Java_com_example_gotestjava_MainActivity_AddSchema
func Java_com_example_gotestjava_MainActivity_AddSchema(env unsafe.Pointer, obj unsafe.Pointer, cSchema unsafe.Pointer) C.int {
	cstr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(cSchema))
	newSchema := C.GoString(cstr)

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

//export Java_com_example_gotestjava_MainActivity_AddDocument
func Java_com_example_gotestjava_MainActivity_AddDocument(env unsafe.Pointer, obj unsafe.Pointer, cColName unsafe.Pointer, cJsonString unsafe.Pointer) C.int {
	cColNameStr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(cColName))
	colName := C.GoString(cColNameStr)
	cJsonStringStr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(cJsonString))
	jsonString := C.GoString(cJsonStringStr)

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

//export Java_com_example_gotestjava_MainActivity_ExecuteQuery
func Java_com_example_gotestjava_MainActivity_ExecuteQuery(env unsafe.Pointer, obj unsafe.Pointer, cQuery unsafe.Pointer) unsafe.Pointer {
	cstr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(cQuery))
	query := C.GoString(cstr)

	ctx := context.Background()

	res := globalNode.DB.ExecRequest(ctx, query)

	// Check for errors in the GQL response, looping through and printing them all
	if len(res.GQL.Errors) > 0 {
		println("Error executing query:")
		for _, err := range res.GQL.Errors {
			println("Error:", err.Error())
		}
		return makeJString((*C.JNIEnv)(env), string("_error_"))
	}

	println("Query executed successfully.")

	// Try to marshall the JSON, and return it
	if res.GQL.Data != nil && len(res.GQL.Data.(map[string]any)) != 0 {
		dataJSON, err := json.Marshal(res.GQL.Data)
		if err != nil {
			println("Error marshalling Data to JSON:", err.Error())
			return makeJString((*C.JNIEnv)(env), string("_error_"))
		}
		println("JSON:")
		println(string(dataJSON))

		return makeJString((*C.JNIEnv)(env), string(dataJSON))
	}

	return makeJString((*C.JNIEnv)(env), string("_error_"))
}

//export Java_com_example_gotestjava_MainActivity_GetString
func Java_com_example_gotestjava_MainActivity_GetString(env unsafe.Pointer, obj unsafe.Pointer, input unsafe.Pointer) unsafe.Pointer {
	// Convert input (jstring) to Go string
	cstr := C.GetStringUTFChars((*C.JNIEnv)(env), (C.jstring)(input))
	goStr := C.GoString(cstr)

	// Process the string
	result := "Hello, " + goStr

	// Convert back to jstring
	jstr := C.NewStringUTF((*C.JNIEnv)(env), C.CString(result))

	// Release the original jstring memory
	C.ReleaseStringUTFChars((*C.JNIEnv)(env), (C.jstring)(input), cstr)

	return unsafe.Pointer(jstr)
}

func main() {}
