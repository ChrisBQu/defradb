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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/immutable"
)

type collectionContextKey struct{}

// Helper function
// Parse the simple (not txn, not identity) options from the C struct into a CollectionFetchOptions object
func parseCollectionOptions(cOptions C.CollectionOptions) client.CollectionFetchOptions {
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
	return options
}

// Helper function
// Set the transaction associated with a given context from the value inside a C.CollectionOptions struct
func setTransactionOfCollectionCommand(ctx context.Context, cOptions C.CollectionOptions) (context.Context, error) {
	TxnIDu64 := uint64(cOptions.tx)
	if TxnIDu64 == 0 {
		return ctx, nil
	}
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return ctx, fmt.Errorf(cerrTxnDoesNotExist, TxnIDu64)
	}
	txn := tx.(datastore.Txn)
	ctx2 := context.WithValue(ctx, transactionContextKey{}, txn)
	return ctx2, nil
}

// Helper function
func getCollectionForCollectionCommand(ctx context.Context, options client.CollectionFetchOptions) (client.Collection, error) {
	// Get the collections that match the options criteria
	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return nil, fmt.Errorf(cerrGettingCollection, err)
	}

	// Make sure only one collection matches the criteria, and select it
	if len(cols) == 0 {
		return nil, fmt.Errorf(cerrNoMatchingCollection)
	}
	if len(cols) > 1 {
		return nil, fmt.Errorf(cerrAmbiguousCollection)
	}
	return cols[0], nil
}

//--------------------------------------------- Exported Functions ---------------------------------------------------------------

//export collectionCreate
func collectionCreate(cJSON *C.char, cIsEncrypted C.int, cEncryptedFields *C.char, cOptions C.CollectionOptions) *C.Result {
	isEncrypted := cIsEncrypted != 0
	jsonString := C.GoString(cJSON)
	jsonBytes := []byte(jsonString)
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Set the correct collection
	foundcol, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	col := foundcol

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the context's collection to the selected one
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

//export collectionDelete
func collectionDelete(cDocID *C.char, cFilter *C.char, cOptions C.CollectionOptions) *C.Result {
	docID := C.GoString(cDocID)
	filter := C.GoString(cFilter)
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Set the correct collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	switch {
	case docID != "":
		ID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(1, fmt.Sprintf("Error: %v", err), "")
		}
		_, err = col.Delete(ctx, ID)
		if err != nil {
			return returnC(1, fmt.Sprintf(cerrDeletingDoc, err), "")
		}
		return returnC(0, "", "")
	case filter != "":
		_, err := col.DeleteWithFilter(ctx, filter)
		if err != nil {
			return returnC(1, fmt.Sprintf(cerrDeletingDoc, err), "")
		}
		return returnC(0, "", "")
	default:
		return returnC(1, cerrNoDocIDOrFilter, "")

	}
}

//export collectionDescribe
func collectionDescribe(cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collections
	cols, err := globalNode.DB.GetCollections(ctx, options)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingCollection, err), "")
	}

	// Get the descriptions
	colDesc := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Definition()
	}

	// Marshall the descriptions into JSON and return
	jsonBytes, err := json.MarshalIndent(colDesc, "", "  ")
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(jsonBytes))
}

//export collectionListDocIDs
func collectionListDocIDs(cOptions C.CollectionOptions) *C.Result {
	ctx := context.Background()
	options := parseCollectionOptions(cOptions)

	// Set the transaction
	newctx, err := setTransactionOfCollectionCommand(ctx, cOptions)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Get the collection
	col, err := getCollectionForCollectionCommand(ctx, options)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Get and return the Doc IDs as a JSON list
	// Note: This is different from the format returned by the CLI, which contains error fields
	docCh, err := col.GetAllDocIDs(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	var docIDs []string
	for doc := range docCh {
		if doc.Err != nil {
			return returnC(1, doc.Err.Error(), "")
		}
		docIDs = append(docIDs, doc.ID.String())
	}
	jsonBytes, err := json.MarshalIndent(docIDs, "", "  ")
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(jsonBytes))
}
