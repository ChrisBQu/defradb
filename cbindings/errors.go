//go:build cgo
// +build cgo

package main

const (
	// Node
	cerrClosingNode            string = "error closing node: %v"
	cerrCreatingStoreDirectory string = "error creating the store directory: %v"
	cerrCreatingNode           string = "error creating node: %v"
	cerrUninitializedNode      string = "error: node is not initialized. Call initNode() first"
	cerrStartingNode           string = "error starting the node: %v"
	cerrStoppedNode            string = "error stopping node: node is not initialized, or was already stopped"
	cerrStoppingNode           string = "error stopping node: %v"
	cerrParsingReplicatorTimes string = "error parsing replicator retry time intervals: %v"
	cerrNegativeReplicatorTime string = "error: negative time intervals are not allowed for replicator retries"

	// Schema
	cerrAddingSchema    string = "error adding schema: %v"
	cerrGettingSchema   string = "error getting schema: %v"
	cerrPatchingSchema  string = "error patching schema: %v"
	cerrSetActiveSchema string = "error setting active version of schema: %v"
	cerrEmptyPatch      string = "patch cannot be empty"

	// Query
	cerrGraphQLResponseEmpty string = "error: graphQL response data is nil or empty"

	// Collection
	cerrGettingCollection    string = "error getting collection: %v"
	cerrCreatingDoc          string = "error creating document: %v"
	cerrInsertingDoc         string = "error inserting document: %v"
	cerrDeletingDoc          string = "error deleting document: %v"
	cerrAmbiguousCollection  string = "error: more than one collection matches the given criteria, could not set context"
	cerrNoMatchingCollection string = "error: no collection matches the given criteria, could not set context"
	cerrNoDocIDOrFilter      string = "error: performing the operation requires a DocID or filter"

	// Index
	cerrInvalidAscensionOrder        string = "invalid ascension order: expected ASC or DESC"
	cerrInvalidIndexFieldDescription string = "invalid or malformed field descriptiona"

	// Txn
	cerrCreatingTxn     string = "error creating transaction: %v"
	cerrCommittingTxn   string = "error committing transaction: %v"
	cerrTxnDoesNotExist string = "error: transaction with ID %v does not exist"

	// Generic
	cerrInvalidLensConfig string = "invalid lens configuration: %v"
	cerrMarshallingJSON   string = "error marshalling JSON: %v"
)
