// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package db provides the implementation of the [client.DB] interface, collection operations,
and related components.
*/
package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/request/graphql"
)

var (
	log = corelog.NewLogger("db")
)

// make sure we match our client interface
var (
	_ client.Collection = (*collection)(nil)
)

const (
	// commandBufferSize is the size of the channel buffer used to handle events.
	commandBufferSize = 100_000
	// eventBufferSize is the size of the channel buffer used to subscribe to events.
	eventBufferSize = 100
)

// DB is the main interface for interacting with the
// DefraDB storage system.
type db struct {
	glock sync.RWMutex

	rootstore  datastore.Rootstore
	multistore datastore.MultiStore

	events *event.Bus

	parser core.Parser

	lensRegistry client.LensRegistry

	// The maximum number of retries per transaction.
	maxTxnRetries immutable.Option[int]

	// The options used to init the database
	options []Option

	// The ID of the last transaction created.
	previousTxnID atomic.Uint64

	// The identity of the current node
	nodeIdentity immutable.Option[identity.Identity]

	// The peer ID and network address information for the current node
	// if network is enabled. The `atomic.Value` should hold a `peer.AddrInfo` struct.
	peerInfo atomic.Value

	// To be able to close the context passed to NewDB on DB close,
	// we need to keep a reference to the cancel function. Otherwise,
	// some goroutines might leak.
	ctxCancel context.CancelFunc

	// The intervals at which to retry replicator failures.
	// For example, this can define an exponential backoff strategy.
	retryIntervals []time.Duration
}

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore datastore.Rootstore,
	lens client.LensRegistry,
	options ...Option,
) (client.DB, error) {
	return newDB(ctx, rootstore, lens, options...)
}

func newDB(
	ctx context.Context,
	rootstore datastore.Rootstore,
	lens client.LensRegistry,
	options ...Option,
) (*db, error) {
	multistore := datastore.MultiStoreFrom(rootstore)

	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}

	opts := defaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	ctx, cancel := context.WithCancel(ctx)

	db := &db{
		rootstore:      rootstore,
		multistore:     multistore,
		lensRegistry:   lens,
		parser:         parser,
		options:        options,
		events:         event.NewBus(commandBufferSize, eventBufferSize),
		ctxCancel:      cancel,
		retryIntervals: opts.RetryIntervals,
	}

	if opts.maxTxnRetries.HasValue() {
		db.maxTxnRetries = opts.maxTxnRetries
	}

	db.nodeIdentity = opts.identity

	if lens != nil {
		lens.Init(db)
	}

	println("Here b.c.a")
	err = db.initialize(ctx)
	if err != nil {
		println("Here b.c.error")
		return nil, err
	}

	sub, err := db.events.Subscribe(event.MergeName, event.PeerInfoName, event.ReplicatorFailureName)
	if err != nil {
		return nil, err
	}
	go db.handleMessages(ctx, sub)
	go db.handleReplicatorRetries(ctx)

	return db, nil
}

// NewTxn creates a new transaction.
func (db *db) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *db) NewConcurrentTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewConcurrentTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// Rootstore returns the root datastore.
func (db *db) Rootstore() datastore.Rootstore {
	return db.rootstore
}

// Blockstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Blockstore() datastore.Blockstore {
	return db.multistore.Blockstore()
}

// Encstore returns the internal enc store which contains encryption key for documents and their fields.
func (db *db) Encstore() datastore.Blockstore {
	return db.multistore.Encstore()
}

// Peerstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Peerstore() datastore.DSReaderWriter {
	return db.multistore.Peerstore()
}

// Headstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Headstore() ds.Read {
	return db.multistore.Headstore()
}

func (db *db) LensRegistry() client.LensRegistry {
	return db.lensRegistry
}

func (db *db) AddPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	return client.AddPolicyResult{}, nil
}

// publishDocUpdateEvent publishes an update event for a document.
// It uses heads iterator to read the document's head blocks directly from the storage, i.e. without
// using a transaction.
func (db *db) publishDocUpdateEvent(ctx context.Context, docID string, collection client.Collection) error {
	headsIterator, err := NewHeadBlocksIterator(ctx, db.multistore.Headstore(), db.Blockstore(), docID)
	if err != nil {
		return err
	}

	for {
		hasValue, err := headsIterator.Next()
		if err != nil {
			return err
		}
		if !hasValue {
			break
		}

		updateEvent := event.Update{
			DocID:      docID,
			Cid:        headsIterator.CurrentCid(),
			SchemaRoot: collection.Schema().Root,
			Block:      headsIterator.CurrentRawBlock(),
		}
		db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
	}
	return nil
}

func (db *db) AddDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddDocActorRelationshipResult, error) {

	return client.AddDocActorRelationshipResult{}, nil
}

func (db *db) DeleteDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteDocActorRelationshipResult, error) {
	return client.DeleteDocActorRelationshipResult{}, nil
}

func (db *db) GetNodeIdentity(context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	if db.nodeIdentity.HasValue() {
		return immutable.Some(db.nodeIdentity.Value().IntoRawIdentity().Public()), nil
	}
	return immutable.None[identity.PublicRawIdentity](), nil
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters.
func (db *db) initialize(ctx context.Context) error {

	println("Context: ", ctx)
	println("Here b.c.a.a")
	db.glock.Lock()
	defer db.glock.Unlock()
	println("Here b.c.a.b")
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	println("Here b.c.a.c")
	exists, err := txn.Systemstore().Has(ctx, ds.NewKey("init"))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and migrations and finish initialization
	println("Here b.c.a.d??")
	if exists {
		err = db.loadSchema(ctx)
		if err != nil {
			return err
		}

		err = db.lensRegistry.ReloadLenses(ctx)
		if err != nil {
			return err
		}

		// The query language types are only updated on successful commit
		// so we must not forget to do so on success regardless of whether
		// we have written to the datastores.
		return txn.Commit(ctx)
	}
	println("Here b.c.a.d")
	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}
	println("Here b.c.a.e")
	err = txn.Systemstore().Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}
	println("Here b.c.a.f")
	return txn.Commit(ctx)
}

// Events returns the events Channel.
func (db *db) Events() *event.Bus {
	return db.events
}

// MaxRetries returns the maximum number of retries per transaction.
// Defaults to `defaultMaxTxnRetries` if not explicitely set
func (db *db) MaxTxnRetries() int {
	if db.maxTxnRetries.HasValue() {
		return db.maxTxnRetries.Value()
	}
	return defaultMaxTxnRetries
}

// PrintDump prints the entire database to console.
func (db *db) PrintDump(ctx context.Context) error {
	return printStore(ctx, db.multistore.Rootstore())
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releasing of resources (i.e.: Badger instance).
func (db *db) Close() {
	log.Info("Closing DefraDB process...")

	db.ctxCancel()

	db.events.Close()

	err := db.rootstore.Close()
	if err != nil {
		log.ErrorE("Failure closing running process", err)
	}

	log.Info("Successfully closed running process")
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) error {
	q := dsq.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(ctx, q)
	if err != nil {
		return err
	}

	for r := range results.Next() {
		log.InfoContext(ctx, "", corelog.Any(r.Key, r.Value))
	}

	return results.Close()
}
