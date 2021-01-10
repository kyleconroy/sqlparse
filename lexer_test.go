// Copyright 2016 PingCAP, Inc.
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
	"fmt"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"github.com/kyleconroy/sqlparse/mysql"
)

func assert(t *testing.T, a, b interface{}) {
	t.Helper()
	if diff := cmp.Diff(a, b); diff != "" {
		t.Errorf("assertion failure: %s", diff)
	}
}

func TestTokenID(t *testing.T) {
	for str, tok := range tokenMap {
		l := NewScanner(str)
		var v yySymType
		if diff := cmp.Diff(tok, l.Lex(&v)); diff != "" {
			t.Errorf("token id: %s", diff)
		}
	}
}

func TestSingleChar(t *testing.T) {
	table := []byte{'|', '&', '-', '+', '*', '/', '%', '^', '~', '(', ',', ')'}
	for _, tok := range table {
		l := NewScanner(string(tok))
		var v yySymType
		if diff := cmp.Diff(int(tok), l.Lex(&v)); diff != "" {
			t.Errorf("single char: %s", diff)
		}
	}
}

type testCaseItem struct {
	str string
	tok int
}

func TestSingleCharOther(t *testing.T) {
	runTest(t, []testCaseItem{
		{"AT", identifier},
		{"?", paramMarker},
		{"PLACEHOLDER", identifier},
		{"=", eq},
		{".", int('.')},
	})
}

func TestAtLeadingIdentifier(t *testing.T) {
	runTest(t, []testCaseItem{
		{"@", singleAtIdentifier},
		{"@''", singleAtIdentifier},
		{"@1", singleAtIdentifier},
		{"@.1_", singleAtIdentifier},
		{"@-1.", singleAtIdentifier},
		{"@~", singleAtIdentifier},
		{"@$", singleAtIdentifier},
		{"@a_3cbbc", singleAtIdentifier},
		{"@`a_3cbbc`", singleAtIdentifier},
		{"@-3cbbc", singleAtIdentifier},
		{"@!3cbbc", singleAtIdentifier},
		{"@@global.test", doubleAtIdentifier},
		{"@@session.test", doubleAtIdentifier},
		{"@@local.test", doubleAtIdentifier},
		{"@@test", doubleAtIdentifier},
		{"@@global.`test`", doubleAtIdentifier},
		{"@@session.`test`", doubleAtIdentifier},
		{"@@local.`test`", doubleAtIdentifier},
		{"@@`test`", doubleAtIdentifier},
	})
}

// func TestUnderscoreCS(t *testing.T) {
// 	var v yySymType
// 	scanner := NewScanner(`_utf8"string"`)
// 	tok := scanner.Lex(&v)
// 	c.Check(tok, Equals, underscoreCS)
// 	tok = scanner.Lex(&v)
// 	c.Check(tok, Equals, stringLit)
//
// 	scanner.reset("N'string'")
// 	tok = scanner.Lex(&v)
// 	c.Check(tok, Equals, underscoreCS)
// 	tok = scanner.Lex(&v)
// 	c.Check(tok, Equals, stringLit)
// }

func TestLiteral(t *testing.T) {
	runTest(t, []testCaseItem{
		{`'''a'''`, stringLit},
		{`''a''`, stringLit},
		{`""a""`, stringLit},
		{`\'a\'`, int('\\')},
		{`\"a\"`, int('\\')},
		{"0.2314", decLit},
		{"1234567890123456789012345678901234567890", decLit},
		{"132.313", decLit},
		{"132.3e231", floatLit},
		{"132.3e-231", floatLit},
		{"001e-12", floatLit},
		{"23416", intLit},
		{"123test", identifier},
		{"123" + string(unicode.ReplacementChar) + "xxx", identifier},
		{"0", intLit},
		{"0x3c26", hexLit},
		{"x'13181C76734725455A'", hexLit},
		{"0b01", bitLit},
		{fmt.Sprintf("t1%c", 0), identifier},
		{"N'some text'", underscoreCS},
		{"n'some text'", underscoreCS},
		{"\\N", null},
		{".*", int('.')},     // `.`, `*`
		{".1_t_1_x", decLit}, // `.1`, `_t_1_x`
		{"9e9e", floatLit},   // 9e9e = 9e9 + e
		{".1e", invalid},
		// Issue #3954
		{".1e23", floatLit},    // `.1e23`
		{".123", decLit},       // `.123`
		{".1*23", decLit},      // `.1`, `*`, `23`
		{".1,23", decLit},      // `.1`, `,`, `23`
		{".1 23", decLit},      // `.1`, `23`
		{".1$23", decLit},      // `.1`, `$23`
		{".1a23", decLit},      // `.1`, `a23`
		{".1e23$23", floatLit}, // `.1e23`, `$23`
		{".1e23a23", floatLit}, // `.1e23`, `a23`
		{".1C23", decLit},      // `.1`, `C23`
		{".1\u0081", decLit},   // `.1`, `\u0081`
		{".1\uff34", decLit},   // `.1`, `\uff34`
		{`b''`, bitLit},
		{`b'0101'`, bitLit},
		{`0b0101`, bitLit},
	})
}

func TestComment(t *testing.T) {
	SpecialCommentsController.Register("test")
	runTest(t, []testCaseItem{
		{"-- select --\n1", intLit},
		{"/*!40101 SET character_set_client = utf8 */;", set},
		{"/* SET character_set_client = utf8 */;", int(';')},
		{"/* some comments */ SELECT ", selectKwd},
		{`-- comment continues to the end of line
SELECT`, selectKwd},
		{`# comment continues to the end of line
SELECT`, selectKwd},
		{"#comment\n123", intLit},
		{"--5", int('-')},
		{"--\nSELECT", selectKwd},
		{"--\tSELECT", 0},
		{"--\r\nSELECT", selectKwd},
		{"--", 0},

		// The odd behavior of '*/' inside conditional comment is the same as
		// that of MySQL.
		{"/*T![unsupported] '*/0 -- ' */", intLit}, // equivalent to 0
		{"/*T![test] '*/0 -- ' */", stringLit},     // equivalent to '*/0 -- '
	})
}

func runTest(t *testing.T, table []testCaseItem) {
	t.Helper()
	var val yySymType
	for _, v := range table {
		l := NewScanner(v.str)
		tok := l.Lex(&val)
		if !cmp.Equal(tok, v.tok) {
			t.Errorf(v.str)
		}
	}
}

func TestScanQuotedIdent(t *testing.T) {
	l := NewScanner("`fk`")
	l.r.peek()
	tok, pos, lit := scanQuotedIdent(l)
	if diff := cmp.Diff(pos.Offset, 0); diff != "" {
		t.Errorf("unexpected pos.Offset: %s", diff)
	}
	if diff := cmp.Diff(tok, quotedIdentifier); diff != "" {
		t.Errorf("unexpected tok: %s", diff)
	}
	if diff := cmp.Diff(lit, "fk"); diff != "" {
		t.Errorf("unexpected lit: %s", diff)
	}
}

func TestScanString(t *testing.T) {
	table := []struct {
		raw    string
		expect string
	}{
		{`' \n\tTest String'`, " \n\tTest String"},
		{`'\x\B'`, "xB"},
		{`'\0\'\"\b\n\r\t\\'`, "\000'\"\b\n\r\t\\"},
		{`'\Z'`, "\x1a"},
		{`'\%\_'`, `\%\_`},
		{`'hello'`, "hello"},
		{`'"hello"'`, `"hello"`},
		{`'""hello""'`, `""hello""`},
		{`'hel''lo'`, "hel'lo"},
		{`'\'hello'`, "'hello"},
		{`"hello"`, "hello"},
		{`"'hello'"`, "'hello'"},
		{`"''hello''"`, "''hello''"},
		{`"hel""lo"`, `hel"lo`},
		{`"\"hello"`, `"hello`},
		{`'disappearing\ backslash'`, "disappearing backslash"},
		{"'한국의中文UTF8およびテキストトラック'", "한국의中文UTF8およびテキストトラック"},
		{"'\\a\x90'", "a\x90"},
		{`"\aèàø»"`, `aèàø»`},
	}
	for _, v := range table {
		l := NewScanner(v.raw)
		tok, pos, lit := l.scan()
		if diff := cmp.Diff(pos.Offset, 0); diff != "" {
			t.Errorf("unexpected pos.Offset: %s", diff)
		}
		if diff := cmp.Diff(tok, stringLit); diff != "" {
			t.Errorf("unexpected tok: %s", diff)
		}
		if diff := cmp.Diff(lit, v.expect); diff != "" {
			t.Errorf("unexpected lit: %s", diff)
		}
	}
}

func TestIdentifier(t *testing.T) {
	replacementString := string(unicode.ReplacementChar) + "xxx"
	table := [][2]string{
		{`哈哈`, "哈哈"},
		{"`numeric`", "numeric"},
		{"\r\n \r \n \tthere\t \n", "there"},
		{`5number`, `5number`},
		{"1_x", "1_x"},
		{"0_x", "0_x"},
		{replacementString, replacementString},
		{"9e", "9e"},
		{"0b", "0b"},
		{"0b123", "0b123"},
		{"0b1ab", "0b1ab"},
		{"0B01", "0B01"},
		{"0x", "0x"},
		{"0x7fz3", "0x7fz3"},
		{"023a4", "023a4"},
		{"9eTSs", "9eTSs"},
		{fmt.Sprintf("t1%cxxx", 0), "t1"},
	}
	l := &Scanner{}
	for _, item := range table {
		l.reset(item[0])
		var v yySymType
		tok := l.Lex(&v)
		if diff := cmp.Diff(tok, identifier); diff != "" {
			t.Errorf("unexpected tok: %s", diff)
		}
		if diff := cmp.Diff(v.ident, item[1]); diff != "" {
			t.Errorf("unexpected v.ident: %s", diff)
		}
	}
}

func TestSpecialComment(t *testing.T) {
	l := NewScanner("/*!40101 select\n5*/")
	tok, pos, lit := l.scan()
	if diff := cmp.Diff(tok, identifier); diff != "" {
		t.Errorf("unexpected tok: %s", diff)
	}
	if diff := cmp.Diff(lit, "select"); diff != "" {
		t.Errorf("unexpected lit: %s", diff)
	}
	if diff := cmp.Diff(pos, Pos{0, 9, 9}); diff != "" {
		t.Errorf("unexpected pos: %s", diff)
	}

	tok, pos, lit = l.scan()
	if diff := cmp.Diff(tok, intLit); diff != "" {
		t.Errorf("unexpected tok: %s", diff)
	}
	if diff := cmp.Diff(lit, "5"); diff != "" {
		t.Errorf("unexpected lit: %s", diff)
	}
	if diff := cmp.Diff(pos, Pos{1, 1, 16}); diff != "" {
		t.Errorf("unexpected pos: %s", diff)
	}
}

func TestFeatureIDsComment(t *testing.T) {
	SpecialCommentsController.Register("auto_rand")
	l := NewScanner("/*T![auto_rand] auto_random(5) */")
	tok, pos, lit := l.scan()
	assert(t, tok, identifier)
	assert(t, lit, "auto_random")
	assert(t, pos, Pos{0, 16, 16})
	tok, pos, lit = l.scan()
	assert(t, tok, int('('))
	tok, pos, lit = l.scan()
	assert(t, lit, "5")
	assert(t, pos, Pos{0, 28, 28})
	tok, pos, lit = l.scan()
	assert(t, tok, int(')'))

	l = NewScanner("/*T![unsupported_feature] unsupported(123) */")
	tok, pos, lit = l.scan()
	assert(t, tok, 0)
}

func TestOptimizerHint(t *testing.T) {
	l := NewScanner("SELECT /*+ BKA(t1) */ 0;")
	tokens := []struct {
		tok   int
		ident string
		pos   int
	}{
		{selectKwd, "SELECT", 0},
		{hintComment, "/*+ BKA(t1) */", 7},
		{intLit, "0", 22},
		{';', ";", 23},
	}
	for i := 0; ; i++ {
		var sym yySymType
		tok := l.Lex(&sym)
		if tok == 0 {
			return
		}
		assert(t, tok, tokens[i].tok)
		assert(t, sym.ident, tokens[i].ident)
		assert(t, sym.offset, tokens[i].pos)
	}
}

func TestOptimizerHintAfterCertainKeywordOnly(t *testing.T) {
	SpecialCommentsController.Register("test")
	tests := []struct {
		input  string
		tokens []int
	}{
		{
			input:  "SELECT /*+ hint */ *",
			tokens: []int{selectKwd, hintComment, '*', 0},
		},
		{
			input:  "UPDATE /*+ hint */",
			tokens: []int{update, hintComment, 0},
		},
		{
			input:  "INSERT /*+ hint */",
			tokens: []int{insert, hintComment, 0},
		},
		{
			input:  "REPLACE /*+ hint */",
			tokens: []int{replace, hintComment, 0},
		},
		{
			input:  "DELETE /*+ hint */",
			tokens: []int{deleteKwd, hintComment, 0},
		},
		{
			input:  "CREATE /*+ hint */",
			tokens: []int{create, hintComment, 0},
		},
		{
			input:  "/*+ hint */ SELECT *",
			tokens: []int{selectKwd, '*', 0},
		},
		{
			input:  "SELECT /* comment */ /*+ hint */ *",
			tokens: []int{selectKwd, hintComment, '*', 0},
		},
		{
			input:  "SELECT * /*+ hint */",
			tokens: []int{selectKwd, '*', 0},
		},
		{
			input:  "SELECT /*T![test] * */ /*+ hint */",
			tokens: []int{selectKwd, '*', 0},
		},
		{
			input:  "SELECT /*T![unsupported] * */ /*+ hint */",
			tokens: []int{selectKwd, hintComment, 0},
		},
		{
			input:  "SELECT /*+ hint1 */ /*+ hint2 */ *",
			tokens: []int{selectKwd, hintComment, '*', 0},
		},
		{
			input:  "SELECT * FROM /*+ hint */",
			tokens: []int{selectKwd, '*', from, 0},
		},
		{
			input:  "`SELECT` /*+ hint */",
			tokens: []int{identifier, 0},
		},
		{
			input:  "'SELECT' /*+ hint */",
			tokens: []int{stringLit, 0},
		},
	}

	for _, tc := range tests {
		scanner := NewScanner(tc.input)
		var sym yySymType
		for i := 0; ; i++ {
			tok := scanner.Lex(&sym)
			if diff := cmp.Diff(tok, tc.tokens[i]); diff != "" {
				t.Errorf("input = [%s], i = %d: %s", tc.input, i, diff)
			}
			if tok == 0 {
				break
			}
		}
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		input  string
		expect uint64
	}{
		{"01000001783", 1000001783},
		{"00001783", 1783},
		{"0", 0},
		{"0000", 0},
		{"01", 1},
		{"10", 10},
	}
	scanner := NewScanner("")
	for _, tc := range tests {
		var v yySymType
		scanner.reset(tc.input)
		tok := scanner.Lex(&v)
		assert(t, tok, intLit)
		switch i := v.item.(type) {
		case int64:
			assert(t, uint64(i), tc.expect)
		case uint64:
			assert(t, i, tc.expect)
		default:
			t.Errorf(tc.input)
		}
	}
}

func TestSQLModeANSIQuotes(t *testing.T) {
	tests := []struct {
		input string
		tok   int
		ident string
	}{
		{`"identifier"`, identifier, "identifier"},
		{"`identifier`", identifier, "identifier"},
		{`"identifier""and"`, identifier, `identifier"and`},
		{`'string''string'`, stringLit, "string'string"},
		{`"identifier"'and'`, identifier, "identifier"},
		{`'string'"identifier"`, stringLit, "string"},
	}
	scanner := NewScanner("")
	scanner.SetSQLMode(mysql.ModeANSIQuotes)
	for _, tc := range tests {
		var v yySymType
		scanner.reset(tc.input)
		tok := scanner.Lex(&v)
		assert(t, tok, tc.tok)
		assert(t, v.ident, tc.ident)
	}

	scanner.reset(`'string' 'string'`)
	var v yySymType

	tok := scanner.Lex(&v)
	assert(t, tok, stringLit)
	assert(t, v.ident, "string")

	tok = scanner.Lex(&v)
	assert(t, tok, stringLit)
	assert(t, v.ident, "string")
}

func TestIllegal(t *testing.T) {
	runTest(t, []testCaseItem{
		{"'", invalid},
		{"'fu", invalid},
		{"'\\n", invalid},
		{"'\\", invalid},
		{fmt.Sprintf("%c", 0), invalid},
		{"`", invalid},
		{`"`, invalid},
		{"@`", invalid},
		{"@'", invalid},
		{`@"`, invalid},
		{"@@`", invalid},
		{"@@global.`", invalid},
	})
}

func TestVersionDigits(t *testing.T) {
	tests := []struct {
		input    string
		min      int
		max      int
		nextChar rune
	}{
		{
			input:    "12345",
			min:      5,
			max:      5,
			nextChar: unicode.ReplacementChar,
		},
		{
			input:    "12345xyz",
			min:      5,
			max:      5,
			nextChar: 'x',
		},
		{
			input:    "1234xyz",
			min:      5,
			max:      5,
			nextChar: '1',
		},
		{
			input:    "123456",
			min:      5,
			max:      5,
			nextChar: '6',
		},
		{
			input:    "1234",
			min:      5,
			max:      5,
			nextChar: '1',
		},
		{
			input:    "",
			min:      5,
			max:      5,
			nextChar: unicode.ReplacementChar,
		},
		{
			input:    "1234567xyz",
			min:      5,
			max:      6,
			nextChar: '7',
		},
		{
			input:    "12345xyz",
			min:      5,
			max:      6,
			nextChar: 'x',
		},
		{
			input:    "12345",
			min:      5,
			max:      6,
			nextChar: unicode.ReplacementChar,
		},
		{
			input:    "1234xyz",
			min:      5,
			max:      6,
			nextChar: '1',
		},
	}

	scanner := NewScanner("")
	for _, tc := range tests {
		scanner.reset(tc.input)
		scanner.scanVersionDigits(tc.min, tc.max)
		nextChar := scanner.r.readByte()
		if diff := cmp.Diff(nextChar, tc.nextChar); diff != "" {
			t.Errorf("input = %s: %s", tc.input, diff)
		}
	}
}

func TestFeatureIDs(t *testing.T) {
	tests := []struct {
		input      string
		featureIDs []string
		nextChar   rune
	}{
		{
			input:      "[feature]",
			featureIDs: []string{"feature"},
			nextChar:   unicode.ReplacementChar,
		},
		{
			input:      "[feature] xx",
			featureIDs: []string{"feature"},
			nextChar:   ' ',
		},
		{
			input:      "[feature1,feature2]",
			featureIDs: []string{"feature1", "feature2"},
			nextChar:   unicode.ReplacementChar,
		},
		{
			input:      "[feature1,feature2,feature3]",
			featureIDs: []string{"feature1", "feature2", "feature3"},
			nextChar:   unicode.ReplacementChar,
		},
		{
			input:      "[id_en_ti_fier]",
			featureIDs: []string{"id_en_ti_fier"},
			nextChar:   unicode.ReplacementChar,
		},
		{
			input:      "[invalid,    whitespace]",
			featureIDs: nil,
			nextChar:   '[',
		},
		{
			input:      "[unclosed_brac",
			featureIDs: nil,
			nextChar:   '[',
		},
		{
			input:      "unclosed_brac]",
			featureIDs: nil,
			nextChar:   'u',
		},
		{
			input:      "[invalid_comma,]",
			featureIDs: nil,
			nextChar:   '[',
		},
		{
			input:      "[,]",
			featureIDs: nil,
			nextChar:   '[',
		},
		{
			input:      "[]",
			featureIDs: nil,
			nextChar:   '[',
		},
	}
	scanner := NewScanner("")
	for _, tc := range tests {
		scanner.reset(tc.input)
		featureIDs := scanner.scanFeatureIDs()
		if diff := cmp.Diff(featureIDs, tc.featureIDs); diff != "" {
			t.Errorf("featureIDs: %s", diff)
		}
		nextChar := scanner.r.readByte()
		if diff := cmp.Diff(nextChar, tc.nextChar); diff != "" {
			t.Errorf("input = %s: %s", tc.input, diff)
		}
	}
}
