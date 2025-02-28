// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/planner"
)

// execRequest executes a request against the database.
func (db *db) execRequest(ctx context.Context, request string, options *client.GQLOptions) *client.RequestResult {

	print("Here exec a.a")
	res := &client.RequestResult{}
	ast, err := db.parser.BuildRequestAST(request)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}
	print("Here exec a.b")
	if db.parser.IsIntrospection(ast) {
		return db.parser.ExecuteIntrospection(request)
	}
	print("Here exec a.c")
	parsedRequest, errors := db.parser.Parse(ast, options)
	if len(errors) > 0 {
		res.GQL.Errors = append(res.GQL.Errors, errors...)
		return res
	}
	print("Here exec a.d")
	pub, err := db.handleSubscription(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}
	print("Here exec a.e")
	if pub != nil {
		res.Subscription = pub
		return res
	}
	print("Here exec a.f")
	txn := mustGetContextTxn(ctx)
	planner := planner.New(ctx, db, txn)
	println("Returned planner:")
	println(planner)
	println("Here exec a.g")
	results, err := planner.RunRequest(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
	}
	print("Here exec a.h")
	res.GQL.Data = results
	return res
}

// ExecIntrospection executes an introspection request against the database.
func (db *db) ExecIntrospection(request string) *client.RequestResult {
	return db.parser.ExecuteIntrospection(request)
}
