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

package parser

import (
	dbsql "database/sql"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	iextract "github.com/kyleconroy/sqlparse/internal/extract"
)

// Add a comment about how this
func CheckCompareReservedWordsWithMySQL(t *testing.T, db *dbsql.DB, dir string) {
	data, err := ioutil.ReadFile(filepath.Join(dir, "parser.y"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	reservedKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordReserved)
	unreservedKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordUnreserved)
	notKeywordTokens := iextract.KeywordsFromTokens(content, iextract.KeywordNot)
	tidbKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordTiDB)

	p := New()
	for _, kw := range reservedKeywords {
		switch kw {
		case "CURRENT_ROLE":
			// special case: we do reserve CURRENT_ROLE but MySQL didn't,
			// and unreservering it causes legit parser conflict.
			continue
		}

		query := "do (select 1 as " + kw + ")"

		var err error

		if _, ok := windowFuncTokenMap[kw]; !ok {
			// for some reason the query does parse even then the keyword is reserved in TiDB.
			_, _, err = p.Parse(query, "", "")
			if !strings.Contains(err.Error(), kw) {
				t.Errorf("error should contain '%s': %s", kw, err)
			}
		}

		_, err = db.Exec(query)
		if !strings.Contains(err.Error(), kw) {
			t.Errorf("MySQL suggests that '%s' should *not* be reserved!", kw)
		}
	}

	for _, kws := range [][]string{unreservedKeywords, notKeywordTokens, tidbKeywords} {
		for _, kw := range kws {
			switch kw {
			case "FUNCTION", // reserved in 8.0.1
				"SEPARATOR": // ?
				continue
			}

			query := "do (select 1 as " + kw + ")"

			stmts, _, err := p.Parse(query, "", "")
			if err != nil {
				t.Errorf("%s: %s", kw, err)
				continue
			}
			if len(stmts) != 1 {
				t.Errorf("%s should have one statement; has %d", kw, len(stmts))
				continue
			}

			// c.Assert(stmts[0], FitsTypeOf, &ast.DoStmt{})
			_, err = db.Exec(query)
			if err != nil {
				t.Errorf("MySQL suggests that '%s' should be reserved!: %s", kw, err)
			}
		}
	}
}
