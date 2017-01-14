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
			Alias(
				"Foo",
				nil,
				BasicType("Int"),
			),
		},
		{
			inputAliasParenBasicType,
			Alias(
				"Foo",
				nil,
				BasicType("Int"),
			),
		},
		{
			inputAliasBasicTypeArg,
			Alias(
				"Foo",
				[]string{"a"},
				BasicType(
					"List",
					BasicType("a"),
				),
			),
		},
		{
			inputAliasBasicTypeArgs,
			Alias(
				"Foo",
				[]string{"a", "b"},
				BasicType(
					"HashMap",
					BasicType("a"),
					BasicType("b"),
				),
			),
		},
		{
			inputAliasRecord,
			Alias(
				"Point",
				nil,
				Record(
					BasicRecordField("x", "Int"),
					BasicRecordField("y", "Int"),
				),
			),
		},
		{
			inputAliasRecordNoFields,
			Alias(
				"Nothing",
				nil,
				Record(),
			),
		},
		{
			inputAliasRecord1Field,
			Alias(
				"X",
				nil,
				Record(
					BasicRecordField("x", "Int"),
				),
			),
		},
		{
			inputAliasRecordNested,
			Alias(
				"Foo",
				nil,
				Record(
					RecordField(
						"x",
						Record(
							BasicRecordField("x1", "Int"),
							BasicRecordField("x2", "Int"),
						),
					),
					BasicRecordField("y", "Int"),
				),
			),
		},
		{
			inputAliasRecordArgs,
			Alias(
				"Foo",
				[]string{"a"},
				Record(
					RecordField(
						"x",
						BasicType("List", BasicType("a")),
					),
				),
			),
		},
		{
			inputAliasTuple,
			Alias(
				"Point",
				nil,
				Tuple(
					BasicType("Int"),
					BasicType("Int"),
				),
			),
		},
		{
			inputAliasTupleArgs,
			Alias(
				"Foo",
				[]string{"a", "b"},
				Tuple(
					BasicType("a"),
					BasicType("b"),
				),
			),
		},
		{
			inputAliasTupleParens,
			Alias(
				"Point",
				nil,
				Tuple(
					BasicType("Int"),
					BasicType("Int"),
				),
			),
		},
		{
			inputAliasFunc,
			Alias(
				"PointMaker",
				nil,
				FuncType(
					BasicType("Int"),
					BasicType("Int"),
					BasicType("Point"),
				),
			),
		},
		{
			inputAliasFuncNested,
			Alias(
				"Foo",
				nil,
				FuncType(
					FuncType(
						BasicType("Int"),
						BasicType("Int"),
						BasicType("String"),
					),
					BasicType("Int"),
					BasicType("Float"),
				),
			),
		},
		{
			inputAliasFuncArgs,
			Alias(
				"Foo",
				[]string{"a", "b"},
				FuncType(
					FuncType(
						BasicType("a"),
						BasicType("a"),
						BasicType("b"),
					),
					BasicType("a"),
					BasicType("b"),
				),
			),
		},
		{
			inputAliasFuncTuple,
			Alias(
				"Foo",
				nil,
				Tuple(
					FuncType(
						BasicType("Int"),
						BasicType("Int"),
					),
					FuncType(
						BasicType("Float"),
						BasicType("Float"),
					),
				),
			),
		},
		{
			inputAliasFuncRecord,
			Alias(
				"Foo",
				nil,
				FuncType(
					Record(
						BasicRecordField("x", "Int"),
						BasicRecordField("y", "Int"),
					),
					Record(
						RecordField(
							"fn",
							FuncType(
								BasicType("Int"),
								BasicType("Int"),
								BasicType("Point"),
							),
						),
					),
					BasicType("Point"),
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
			Union(
				"Foo",
				nil,
				Constructor("A"),
			),
		},
		{
			inputUnionNames,
			Union(
				"Cmp",
				nil,
				Constructor("Lt"),
				Constructor("Eq"),
				Constructor("Gt"),
			),
		},
		{
			inputUnionArgs,
			Union(
				"Cmp",
				[]string{"a"},
				Constructor("Lt", BasicType("a")),
				Constructor("Eq", BasicType("a")),
				Constructor("Gt", BasicType("a")),
			),
		},
		{
			inputUnionRecords,
			Union(
				"Foo",
				[]string{"a", "b"},
				Constructor(
					"A",
					Record(
						BasicRecordField("a", "a"),
					),
				),
				Constructor(
					"B",
					Record(
						BasicRecordField("b", "b"),
						BasicRecordField("c", "String"),
					),
				),
				Constructor(
					"C",
					BasicType("List", BasicType("Int")),
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

const (
	inputLiteral    = `foo = 5`
	inputLiteralAnn = `foo : Int
foo = 5`
	inputOperator    = `(::) a b = 5`
	inputOperatorAnn = `(::) : Int -> Int -> Int
(::) a b = 5`
)

func TestParseDefinition(t *testing.T) {
	cases := []struct {
		input  string
		assert declAssert
	}{
		{
			inputLiteral,
			Definition("foo", nil, nil, Literal(ast.Int, "5")),
		},
		{
			inputLiteralAnn,
			Definition(
				"foo",
				TypeAnnotation(BasicType("Int")),
				nil,
				Literal(ast.Int, "5"),
			),
		},
		{
			inputOperator,
			Definition(
				"::",
				nil,
				Patterns(VarPattern("a"), VarPattern("b")),
				Literal(ast.Int, "5"),
			),
		},
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, "", false)

			p := stringParser(c.input)
			c.assert(t, p.parseDefinition())
		}()
	}
}

func TestParsePattern(t *testing.T) {
	cases := []struct {
		input  string
		assert patternAssert
	}{
		{
			`_`,
			AnythingPattern,
		},
		{
			`"foo"`,
			LiteralPattern(ast.String, `"foo"`),
		},
		{
			`True`,
			CtorPattern("True"),
		},
		{
			`a`,
			VarPattern("a"),
		},
		{
			`Just 42`,
			CtorPattern("Just", LiteralPattern(ast.Int, "42")),
		},
		{
			`(Just 42)`,
			CtorPattern("Just", LiteralPattern(ast.Int, "42")),
		},
		{
			`(a, b, _)`,
			TuplePattern(
				VarPattern("a"),
				VarPattern("b"),
				AnythingPattern,
			),
		},
		{
			`{a, b, c}`,
			RecordPattern(
				VarPattern("a"),
				VarPattern("b"),
				VarPattern("c"),
			),
		},
		{
			`Just 42 as m`,
			AliasPattern(
				CtorPattern("Just", LiteralPattern(ast.Int, "42")),
				"m",
			),
		},
		{
			`[1, 2, _]`,
			ListPattern(
				LiteralPattern(ast.Int, "1"),
				LiteralPattern(ast.Int, "2"),
				AnythingPattern,
			),
		},
		{
			`a::b::_`,
			CtorPattern(
				"::",
				VarPattern("a"),
				CtorPattern(
					"::",
					VarPattern("b"),
					AnythingPattern,
				),
			),
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(st *testing.T) {
			c.input += "\n"
			defer assertEOF(t, "", false)

			p := stringParser(c.input)
			c.assert(st, p.parsePattern(true))
		})
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
