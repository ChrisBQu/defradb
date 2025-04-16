// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package issues

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func Test_CastIntToString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(includes: [{field: "custom"}, {field: "tags"}]) {
						name: String 
						custom: JSON 
						tags: [String]
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Chris",
					"custom": map[string]any{
						"numbers": []int{},
					},
					"tags": []string{"tag1", "tag2", "tag3"},
				},
			},
			testUtils.Request{
				Request: `
					query {
						User {
							name
							tags
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"tags": []string{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
