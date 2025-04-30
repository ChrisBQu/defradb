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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/encryption"
)

type collectionContextKey struct{}

//export passInStruct
func passInStruct(cOptions C.CollectionOptions) *C.Result {
	finalname := C.GoString(cOptions.name)

	return returnC(0, "", fmt.Sprintf("You passed in: %v", finalname))
}

//export createDocument
func createDocument(cJSON *C.char, cIsEncrypted C.int, cEncryptedFields *C.char, cOptions C.CollectionOptions) *C.Result {
	isEncrypted := cIsEncrypted != 0

	jsonString := C.GoString(cJSON)
	jsonBytes := []byte(jsonString)
	ctx := context.Background()

	// Parse the simple (not txn, not identity) options into a CollectionFetchOptions object
	versionID := C.GoString(cOptions.version)
	schemaRoot := C.GoString(cOptions.schema)
	name := C.GoString(cOptions.name)
	getInactive := cOptions.getInactive != 0
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

	// Get the collections that match the criteria
	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingCollection, err), "")
	}

	// Make sure only one collection matches the criteria
	if len(cols) == 0 {
		return returnC(1, cerrNoMatchingCollection, "")
	}
	if len(cols) > 1 {
		return returnC(1, cerrAmbiguousCollection, "")
	}

	// Set the context's collection to the selected one
	col := cols[0]
	ctx = context.WithValue(ctx, collectionContextKey{}, col)

	// Set the encryption
	raw := C.GoString(cEncryptedFields)
	var encryptFields []string
	if raw != "" {
		for _, f := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				encryptFields = append(encryptFields, trimmed)
			}
		}
	}
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted, encryptFields)

	// Create the document(s)
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
		return returnC(1, fmt.Sprintf(cerrGettingCollection, err), "")
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
		return returnC(1, fmt.Sprintf(cerrGettingCollection, err), "")
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
