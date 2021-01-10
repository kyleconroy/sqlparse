// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

//+build integration

// This file ensures that the set of reserved keywords is the same as that of
// MySQL. To run:
//
//  1. Set up a MySQL server listening at 127.0.0.1:3306 using root and password `mysecretpassword`
//  2. Run this test with:
//
//		go test -tags integration ./...
package integration

import (
	dbsql "database/sql"
	"testing"

	// needed to connect to MySQL
	_ "github.com/go-sql-driver/mysql"

	parser "github.com/kyleconroy/sqlparse"
)

func TestCompareReservedWordsWithMySQL(t *testing.T) {
	db, err := dbsql.Open("mysql", "root@tcp(127.0.0.1:3306)/")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	parser.CheckCompareReservedWordsWithMySQL(t, db)
}
