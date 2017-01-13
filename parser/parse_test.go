package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/stretchr/testify/require"
)

func TestParseModule(t *testing.T) {
	require := require.New(t)
	cases := []struct {
		input   string
		ok, eof bool
		module  string
		exposed [][]string
	}{
		{"module Foo", true, false, "Foo", nil},
		{"module foo", false, false, "", nil},
		{"bar Foo", false, false, "", nil},
		{"module Foo.Bar", true, false, "Foo.Bar", nil},
		{"module Foo.Bar.Baz", true, false, "Foo.Bar.Baz", nil},
		{"module Foo exposing", false, true, "Foo", nil},
		{"module Foo exposing ()", false, false, "Foo", nil},
		{"module Foo exposing (..)", true, false, "Foo", [][]string{{".."}}},
		{"module Foo exposing (foo)", true, false, "Foo", [][]string{{"foo"}}},
		{"module Foo exposing (foo(..))", false, false, "", nil},
		{"module Foo exposing (Foo(foo, Bar))", false, false, "", nil},
		{"module Foo exposing (foo, bar)", true, false, "Foo", [][]string{{"foo"}, {"bar"}}},
		{"module Foo exposing (foo, bar, baz)", true, false, "Foo", [][]string{{"foo"}, {"bar"}, {"baz"}}},
		{"module Foo exposing (foo, (:>), baz)", true, false, "Foo", [][]string{{"foo"}, {":>"}, {"baz"}}},
		{"module Foo exposing ((:>), (:>), (:>))", true, false, "Foo", [][]string{{":>"}, {":>"}, {":>"}}},
		{"module Foo exposing (foo, Bar(..), Baz(A, B, C))", true, false, "Foo", [][]string{
			{"foo"},
			{"Bar", ".."},
			{"Baz", "A", "B", "C"},
		}},
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, c.input, c.eof)

			p := stringParser(c.input)
			decl := p.parseModule()

			if c.ok {
				var exposed [][]string
				if decl.Exposing != nil {
					for _, e := range decl.Exposing.Idents {
						var exp = []string{e.Name}
						if e.Exposing != nil {
							for _, e := range e.Exposing.Idents {
								exp = append(exp, e.Name)
							}
						}
						exposed = append(exposed, exp)
					}
				}

				require.Equal(0, len(p.errors), c.input)
				require.Equal(c.module, decl.Name.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
			} else {
				require.NotEqual(0, len(p.errors), c.input)
			}
		}()
	}
}

func TestParseImport(t *testing.T) {
	require := require.New(t)
	cases := []struct {
		input   string
		ok, eof bool
		module  string
		alias   string
		exposed [][]string
	}{
		{"import Foo", true, false, "Foo", "", nil},
		{"import foo", false, false, "", "", nil},
		{"bar Foo", false, false, "", "", nil},
		{"import Foo.Bar", true, false, "Foo.Bar", "", nil},
		{"import Foo.Bar.Baz", true, false, "Foo.Bar.Baz", "", nil},
		{"import Foo.Bar.Baz as Foo", true, false, "Foo.Bar.Baz", "Foo", nil},
		{"import Foo exposing", false, true, "Foo", "", nil},
		{"import Foo exposing ()", false, false, "Foo", "", nil},
		{"import Foo exposing (..)", true, false, "Foo", "", [][]string{{".."}}},
		{"import Foo as Bar exposing (..)", true, false, "Foo", "Bar", [][]string{{".."}}},
		{"import foo as bar exposing (..)", false, false, "", "", nil},
		{"import Foo exposing (foo)", true, false, "Foo", "", [][]string{{"foo"}}},
		{"import Foo exposing (foo, bar)", true, false, "Foo", "", [][]string{{"foo"}, {"bar"}}},
		{"import Foo exposing (foo, bar, baz)", true, false, "Foo", "", [][]string{{"foo"}, {"bar"}, {"baz"}}},
		{"import Foo exposing (foo, (:>), baz)", true, false, "Foo", "", [][]string{{"foo"}, {":>"}, {"baz"}}},
		{"import Foo exposing ((:>), (:>), (:>))", true, false, "Foo", "", [][]string{{":>"}, {":>"}, {":>"}}},
		{"import Foo exposing (bar(..))", false, false, "", "", nil},
		{"import Foo exposing (foo, Bar(..), Baz(A, B, C))", true, false, "Foo", "", [][]string{
			{"foo"},
			{"Bar", ".."},
			{"Baz", "A", "B", "C"},
		}},
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, c.input, c.eof)

			p := stringParser(c.input)
			decl := p.parseImport()

			if c.ok {
				var exposed [][]string
				if decl.Exposing != nil {
					for _, e := range decl.Exposing.Idents {
						var exp = []string{e.Name}
						if e.Exposing != nil {
							for _, e := range e.Exposing.Idents {
								exp = append(exp, e.Name)
							}
						}
						exposed = append(exposed, exp)
					}
				}

				require.Equal(0, len(p.errors), c.input)
				require.Equal(c.module, decl.Module.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
				if c.alias != "" {
					require.NotNil(decl.Alias, c.input)
					require.Equal(c.alias, decl.Alias.Name, c.input)
				}
			} else {
				require.NotEqual(0, len(p.errors), c.input)
			}
		}()
	}
}

func TestParseInfixDecl(t *testing.T) {
	cases := []struct {
		input    string
		assoc    ast.Associativity
		op       string
		priority int
		ok       bool
		eof      bool
	}{
		{"infixr 4 ?", ast.RightAssoc, "?", 4, true, false},
		{"infixl 4 ?", ast.LeftAssoc, "?", 4, true, false},
		{"infix 4 ?", ast.NonAssoc, "?", 4, true, false},
		{"infixl 0 ?", ast.LeftAssoc, "?", 0, true, false},
		{"infixl 4 foo", ast.LeftAssoc, "", 0, false, false},
		{"infixl \"a\" ?", ast.LeftAssoc, "", 0, false, false},
		{"infixl ? 5", ast.LeftAssoc, "", 0, false, false},
		{"infixl ?", ast.LeftAssoc, "", 0, false, true},
		{"infixl -1 ?", ast.LeftAssoc, "", 0, false, false},
		{"infixl 10 ?", ast.LeftAssoc, "", 0, false, false},
		{"infixl 20 ?", ast.LeftAssoc, "", 0, false, false},
	}

	require := require.New(t)
	for _, c := range cases {
		func() {
			defer assertEOF(t, c.input, c.eof)

			p := stringParser(c.input)
			decl := p.parseInfixDecl().(*ast.InfixDecl)
			if c.ok {
				require.Equal(c.assoc, decl.Assoc, c.input)
				require.Equal(c.op, decl.Op.Name, c.input)
				p, err := strconv.Atoi(decl.Priority.Value)
				require.Nil(err, c.input)
				require.Equal(c.priority, p, c.input)
			} else {
				require.NotEqual(0, len(p.errors), c.input)
			}
		}()
	}
}

const inputAliasSimpleType = `
type alias Foo = Int
`

const inputAliasParenBasicType = `
type alias Foo = (Int)
`

const inputAliasBasicTypeArg = `
type alias Foo a = List a
`

const inputAliasBasicTypeArgs = `
type alias Foo a b = HashMap a b
`

const inputAliasRecord = `
type alias Point = {x: Int, y: Int}
`

const inputAliasRecordNoFields = `
type alias Nothing = {}
`

const inputAliasRecord1Field = `
type alias X = {x: Int}
`

const inputAliasRecordNested = `
type alias Foo = {x: {x1: Int, x2: Int}, y: Int}
`

const inputAliasRecordArgs = `
type alias Foo a = {x: List a}
`

const inputAliasTuple = `
type alias Point = (Int, Int)
`

const inputAliasTupleArgs = `
type alias Foo a b = (a, b)
`

const inputAliasTupleParens = `
type alias Point = (((Int), (Int)))
`

const inputAliasFunc = `
type alias PointMaker = Int -> Int -> Point
`

const inputAliasFuncNested = `
type alias Foo = (Int -> Int -> String) -> Int -> Float
`

const inputAliasFuncArgs = `
type alias Foo a b = (a -> a -> b) -> a -> b
`

const inputAliasFuncTuple = `
type alias Foo = (Int -> Int, Float -> Float)
`

const inputAliasFuncRecord = `
type alias Foo = {x: Int, y: Int} -> {fn: Int -> Int -> Point} -> Point
`

func TestParseTypeAlias(t *testing.T) {
	cases := []struct {
		input  string
		assert declAssert
	}{
		{
			inputAliasSimpleType,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertBasicType("Int"),
			),
		},
		{
			inputAliasParenBasicType,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertBasicType("Int"),
			),
		},
		{
			inputAliasBasicTypeArg,
			assertAlias(
				assertName("Foo"),
				assertArgs("a"),
				assertBasicType(
					"List",
					assertBasicType("a"),
				),
			),
		},
		{
			inputAliasBasicTypeArgs,
			assertAlias(
				assertName("Foo"),
				assertArgs("a", "b"),
				assertBasicType(
					"HashMap",
					assertBasicType("a"),
					assertBasicType("b"),
				),
			),
		},
		{
			inputAliasRecord,
			assertAlias(
				assertName("Point"),
				assertNoArgs,
				assertRecord(
					assertBasicRecordField("x", "Int"),
					assertBasicRecordField("y", "Int"),
				),
			),
		},
		{
			inputAliasRecordNoFields,
			assertAlias(
				assertName("Nothing"),
				assertNoArgs,
				assertRecord(),
			),
		},
		{
			inputAliasRecord1Field,
			assertAlias(
				assertName("X"),
				assertNoArgs,
				assertRecord(
					assertBasicRecordField("x", "Int"),
				),
			),
		},
		{
			inputAliasRecordNested,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertRecord(
					assertRecordField(
						"x",
						assertRecord(
							assertBasicRecordField("x1", "Int"),
							assertBasicRecordField("x2", "Int"),
						),
					),
					assertBasicRecordField("y", "Int"),
				),
			),
		},
		{
			inputAliasRecordArgs,
			assertAlias(
				assertName("Foo"),
				assertArgs("a"),
				assertRecord(
					assertRecordField(
						"x",
						assertBasicType("List", assertBasicType("a")),
					),
				),
			),
		},
		{
			inputAliasTuple,
			assertAlias(
				assertName("Point"),
				assertNoArgs,
				assertTuple(
					assertBasicType("Int"),
					assertBasicType("Int"),
				),
			),
		},
		{
			inputAliasTupleArgs,
			assertAlias(
				assertName("Foo"),
				assertArgs("a", "b"),
				assertTuple(
					assertBasicType("a"),
					assertBasicType("b"),
				),
			),
		},
		{
			inputAliasTupleParens,
			assertAlias(
				assertName("Point"),
				assertNoArgs,
				assertTuple(
					assertBasicType("Int"),
					assertBasicType("Int"),
				),
			),
		},
		{
			inputAliasFunc,
			assertAlias(
				assertName("PointMaker"),
				assertNoArgs,
				assertFuncType(
					assertBasicType("Int"),
					assertBasicType("Int"),
					assertBasicType("Point"),
				),
			),
		},
		{
			inputAliasFuncNested,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertFuncType(
					assertFuncType(
						assertBasicType("Int"),
						assertBasicType("Int"),
						assertBasicType("String"),
					),
					assertBasicType("Int"),
					assertBasicType("Float"),
				),
			),
		},
		{
			inputAliasFuncArgs,
			assertAlias(
				assertName("Foo"),
				assertArgs("a", "b"),
				assertFuncType(
					assertFuncType(
						assertBasicType("a"),
						assertBasicType("a"),
						assertBasicType("b"),
					),
					assertBasicType("a"),
					assertBasicType("b"),
				),
			),
		},
		{
			inputAliasFuncTuple,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertTuple(
					assertFuncType(
						assertBasicType("Int"),
						assertBasicType("Int"),
					),
					assertFuncType(
						assertBasicType("Float"),
						assertBasicType("Float"),
					),
				),
			),
		},
		{
			inputAliasFuncRecord,
			assertAlias(
				assertName("Foo"),
				assertNoArgs,
				assertFuncType(
					assertRecord(
						assertBasicRecordField("x", "Int"),
						assertBasicRecordField("y", "Int"),
					),
					assertRecord(
						assertRecordField(
							"fn",
							assertFuncType(
								assertBasicType("Int"),
								assertBasicType("Int"),
								assertBasicType("Point"),
							),
						),
					),
					assertBasicType("Point"),
				),
			),
		},
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, "", false)

			p := stringParser(c.input)
			c.assert(t, p.parseTypeDecl())
		}()
	}
}

const inputUnionOne = `
type Foo = A
`

const inputUnionNames = `
type Cmp = Lt | Eq | Gt
`

const inputUnionArgs = `
type Cmp a = Lt a | Eq a | Gt a
`

const inputUnionRecords = `
type Foo a b = A {a: a} | B {b: b, c: String} | C (List Int)
`

func TestParseTypeUnion(t *testing.T) {
	cases := []struct {
		input  string
		assert declAssert
	}{
		{
			inputUnionOne,
			assertUnion(
				assertName("Foo"),
				assertNoArgs,
				assertConstructor("A"),
			),
		},
		{
			inputUnionNames,
			assertUnion(
				assertName("Cmp"),
				assertNoArgs,
				assertConstructor("Lt"),
				assertConstructor("Eq"),
				assertConstructor("Gt"),
			),
		},
		{
			inputUnionArgs,
			assertUnion(
				assertName("Cmp"),
				assertArgs("a"),
				assertConstructor("Lt", assertBasicType("a")),
				assertConstructor("Eq", assertBasicType("a")),
				assertConstructor("Gt", assertBasicType("a")),
			),
		},
		{
			inputUnionRecords,
			assertUnion(
				assertName("Foo"),
				assertArgs("a", "b"),
				assertConstructor(
					"A",
					assertRecord(
						assertBasicRecordField("a", "a"),
					),
				),
				assertConstructor(
					"B",
					assertRecord(
						assertBasicRecordField("b", "b"),
						assertBasicRecordField("c", "String"),
					),
				),
				assertConstructor(
					"C",
					assertBasicType("List", assertBasicType("Int")),
				),
			),
		},
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, "", false)

			p := stringParser(c.input)
			c.assert(t, p.parseTypeDecl())
		}()
	}
}

func assertEOF(t *testing.T, input string, eof bool) {
	if r := recover(); r != nil {
		switch r.(type) {
		case bailout:
			if !eof {
				require.FailNow(t, "unexpected bailout", input)
			}
		default:
			panic(r)
		}
	}
}

func stringParser(str string) *parser {
	scanner := scanner.New("test", strings.NewReader(str))
	go scanner.Run()
	var p = new(parser)
	p.init("test", scanner)
	return p
}
