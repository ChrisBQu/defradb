//go:build cgo
// +build cgo

package main

/*
#include "result.h"
*/
import "C"

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

//export createDocument
func createDocument(cCollection *C.char, cJSON *C.char, cIdentity *C.char, cIsEncrypted C.int, cEncryptedFields *C.char) *C.Result {
	colName := C.GoString(cCollection)
	isEncrypted := cIsEncrypted != 0
	encryptFields := strings.Split(C.GoString(cEncryptedFields), ",")
	jsonString := C.GoString(cJSON)
	jsonBytes := []byte(jsonString)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingCollectionByName, err), "")
	}

	// Set the encryption
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted, encryptFields)

	// If the collection exists, then try to create a new document
	// If this doesn't work, then fail out

	if client.IsJSONArray(jsonBytes) {
		docs, err := client.NewDocsFromJSON(jsonBytes, col.Definition())
		if err != nil {
			return returnC(1, fmt.Sprintf("Error: %v", err), "")
		}
		err2 := col.CreateMany(ctx, docs)
		if err2 != nil {
			return returnC(1, fmt.Sprintf("Error: %v", err2), "")
		}
		return returnC(0, "", "")
	}

	doc, err := client.NewDocFromJSON(jsonBytes, col.Definition())
	if err != nil {
		return returnC(1, fmt.Sprintf("Error: %v", err), "")
	}
	err2 := col.Create(ctx, doc)
	if err2 != nil {
		return returnC(1, fmt.Sprintf("Error: %v", err2), "")
	}

	return returnC(0, "", "")
}

//---------------------------- Legacy

//export addDocument
func addDocument(cCollection *C.char, cJSON *C.char) *C.Result {
	colName := C.GoString(cCollection)
	jsonString := C.GoString(cJSON)
	ctx := context.Background()

	// Check if the collection exists, and if it doesn't fail
	col, err := globalNode.DB.GetCollectionByName(ctx, colName)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingCollectionByName, err), "")
	}

	// If the collection exists, then try to create a new document
	// If this doesn't work, then fail out
	doc, err := client.NewDocFromJSON([]byte(jsonString), col.Definition())
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingDoc, err), "")
	}

	// Now we will try to insert the document into the collection
	err = col.Create(ctx, doc)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrInsertingDoc, err), "")
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
		return returnC(1, fmt.Sprintf(cerrGettingCollectionByName, err), "")
	}

	// Try to get a document ID from the string passed in
	docIDStruct, err := client.NewDocIDFromString(docID)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingDoc, err), "")
	}

	// If the collection exists, then try to delete the document from it
	success, err := col.Delete(ctx, docIDStruct)
	if success != true {
		return returnC(1, fmt.Sprintf(cerrDeletingDoc, err), "")
	}

	return returnC(0, "", "")
}
