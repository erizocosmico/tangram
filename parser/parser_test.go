package parser

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/source"
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

			p := stringParser(t, c.input)
			defer p.sess.Emit()
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
				n, ok := decl.Name.(fmt.Stringer)
				require.True(ok, "expected module name to be stringer")
				require.Equal(c.module, n.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
			} else {
				require.False(p.sess.IsOK(), c.input)
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

			p := stringParser(t, c.input)
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
				mod, ok := decl.Module.(fmt.Stringer)
				require.True(ok, "expected module name to be stringer")
				require.Equal(c.module, mod.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
				if c.alias != "" {
					require.NotNil(decl.Alias, c.input)
					require.Equal(c.alias, decl.Alias.Name, c.input)
				}
			} else {
				require.False(p.sess.IsOK(), c.input)
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

			p := stringParser(t, c.input)
			decl := p.parseInfixDecl().(*ast.InfixDecl)
			if c.ok {
				require.Equal(c.assoc, decl.Assoc, c.input)
				require.Equal(c.op, decl.Op.Name, c.input)
				p, err := strconv.Atoi(decl.Priority.Value)
				require.Nil(err, c.input)
				require.Equal(c.priority, p, c.input)
			} else {
				require.False(p.sess.IsOK(), c.input)
			}
		}()
	}
}

const inputAliasSimpleType = `
type alias Foo = Int
`

const inputAliasSimpleTypeSelector = `
type alias Foo = List.List
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
			inputAliasSimpleTypeSelector,
			Alias(
				"Foo",
				nil,
				SelectorType(Selector("List", "List")),
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

			p := stringParser(t, c.input)
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

			p := stringParser(t, c.input)
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

			p := stringParser(t, c.input)
			c.assert(t, p.parseDefinition())
		}()
	}
}

func TestParseDestructuringAssignment(t *testing.T) {
	cases := []struct {
		input  string
		assert declAssert
	}{
		{
			`( a, b ) = ( 1, 2 )`,
			Destructuring(
				TuplePattern(
					VarPattern("a"),
					VarPattern("b"),
				),
				TupleLiteral(
					Literal(ast.Int, "1"),
					Literal(ast.Int, "2"),
				),
			),
		},
		{
			`{ x, y } = { x = 1, y = 2 }`,
			Destructuring(
				RecordPattern(
					VarPattern("x"),
					VarPattern("y"),
				),
				RecordLiteral(
					FieldAssign("x", Literal(ast.Int, "1")),
					FieldAssign("y", Literal(ast.Int, "2")),
				),
			),
		},
		{
			`_ = 2`,
			Destructuring(
				AnythingPattern,
				Literal(ast.Int, "2"),
			),
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			defer assertEOF(t, "", false)

			p := stringParser(t, c.input)
			c.assert(t, p.parseDestructuringAssignment())
		})
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

			p := stringParser(t, c.input)
			c.assert(st, p.parsePattern(true))
		})
	}
}

func TestParseType(t *testing.T) {
	cases := []struct {
		input  string
		assert typeAssert
	}{
		{
			"List.List",
			SelectorType(Selector("List", "List")),
		},
		{
			"Map.Map Foo.Bar List.List",
			SelectorType(
				Selector("Map", "Map"),
				SelectorType(Selector("Foo", "Bar")),
				SelectorType(Selector("List", "List")),
			),
		},
		{
			"HashMap (Foo a) (List Int)",
			BasicType("HashMap",
				BasicType("Foo", BasicType("a")),
				BasicType("List", BasicType("Int")),
			),
		},
		// TODO(erizocosmico): improve this tests cases and relieve pressure
		// from ParseTypeUnion and ParseTypeAlias
	}

	for _, c := range cases {
		func() {
			defer assertEOF(t, "", false)

			// the space here is because a type can not be at the start of a
			// line
			p := stringParser(t, " "+c.input)
			typ := p.parseType()
			if c.assert == nil {
				require.Nil(t, typ, "expected type to be nil")
			} else {
				c.assert(t, typ)
			}
		}()
	}
}

func TestParseExpr(t *testing.T) {
	cases := []struct {
		input  string
		assert exprAssert
	}{
		{`5`, Literal(ast.Int, "5")},
		{`"hello world"`, Literal(ast.String, `"hello world"`)},
		{`True`, Literal(ast.Bool, `True`)},
		{`False`, Literal(ast.Bool, `False`)},
		{`3.1416`, Literal(ast.Float, `3.1416`)},
		{`'a'`, Literal(ast.Char, `'a'`)},
		{`()`, TupleLiteral()},
		{`[]`, ListLiteral()},
		{
			`(1, 2, 3)`,
			TupleLiteral(
				Literal(ast.Int, "1"),
				Literal(ast.Int, "2"),
				Literal(ast.Int, "3"),
			),
		},
		{
			`[1, 2, 3]`,
			ListLiteral(
				Literal(ast.Int, "1"),
				Literal(ast.Int, "2"),
				Literal(ast.Int, "3"),
			),
		},
		{
			`((1, 2), (2, 3))`,
			TupleLiteral(
				TupleLiteral(
					Literal(ast.Int, "1"),
					Literal(ast.Int, "2"),
				),
				TupleLiteral(
					Literal(ast.Int, "2"),
					Literal(ast.Int, "3"),
				),
			),
		},
		{
			`[[1, 2], [2, 3]]`,
			ListLiteral(
				ListLiteral(
					Literal(ast.Int, "1"),
					Literal(ast.Int, "2"),
				),
				ListLiteral(
					Literal(ast.Int, "2"),
					Literal(ast.Int, "3"),
				),
			),
		},
		{`(,,)`, TupleCtor(3)},
		{
			`{ a = 1, b = [ 1, 2 ], c = { x = 1, y = 2 } }`,
			RecordLiteral(
				FieldAssign("a", Literal(ast.Int, "1")),
				FieldAssign("b", ListLiteral(
					Literal(ast.Int, "1"),
					Literal(ast.Int, "2"),
				)),
				FieldAssign("c", RecordLiteral(
					FieldAssign("x", Literal(ast.Int, "1")),
					FieldAssign("y", Literal(ast.Int, "2")),
				)),
			),
		},
		{
			`{ point | x = 5, y = 2 }`,
			RecordUpdate(
				"point",
				FieldAssign("x", Literal(ast.Int, "5")),
				FieldAssign("y", Literal(ast.Int, "2")),
			),
		},
		{
			`\a (x, y) {z, d} _-> \k-> 5`,
			Lambda(
				Patterns(
					VarPattern("a"),
					TuplePattern(
						VarPattern("x"),
						VarPattern("y"),
					),
					RecordPattern(
						VarPattern("z"),
						VarPattern("d"),
					),
					AnythingPattern,
				),
				Lambda(
					Patterns(VarPattern("k")),
					Literal(ast.Int, "5"),
				),
			),
		},
		{
			`let
				foo = 5

				bar a b = 6

				( a, b ) = qux

				{ x, y } = baz

				mux : Int
				mux = 7

				_ = ignored
			in
				5`,
			Let(
				Literal(ast.Int, "5"),
				Definition("foo", nil, nil, Literal(ast.Int, "5")),
				Definition("bar", nil,
					Patterns(
						VarPattern("a"),
						VarPattern("b"),
					),
					Literal(ast.Int, "6"),
				),
				Destructuring(
					TuplePattern(
						VarPattern("a"),
						VarPattern("b"),
					),
					Identifier("qux"),
				),
				Destructuring(
					RecordPattern(
						VarPattern("x"),
						VarPattern("y"),
					),
					Identifier("baz"),
				),
				Definition("mux",
					TypeAnnotation(BasicType("Int")),
					nil,
					Literal(ast.Int, "7"),
				),
				Destructuring(
					AnythingPattern,
					Identifier("ignored"),
				),
			),
		},
		{
			`f <| g a b`,
			BinaryExpr(
				"<|",
				Identifier("f"),
				FuncApp(
					Identifier("g"),
					Identifier("a"),
					Identifier("b"),
				),
			),
		},
		{
			`a + b + c`,
			BinaryExpr(
				"+",
				BinaryExpr(
					"+",
					Identifier("a"),
					Identifier("b"),
				),
				Identifier("c"),
			),
		},
		{
			`a + b * c - d`,
			BinaryExpr(
				"-",
				BinaryExpr(
					"+",
					Identifier("a"),
					BinaryExpr(
						"*",
						Identifier("b"),
						Identifier("c"),
					),
				),
				Identifier("d"),
			),
		},
		{
			`map ls fn`,
			FuncApp(
				Identifier("map"),
				Identifier("ls"),
				Identifier("fn"),
			),
		},
		{
			`(f a b) c d`,
			FuncApp(
				Parens(
					FuncApp(
						Identifier("f"),
						Identifier("a"),
						Identifier("b"),
					),
				),
				Identifier("c"),
				Identifier("d"),
			),
		},
		{
			`a + -b * c - d`,
			BinaryExpr(
				"-",
				BinaryExpr(
					"+",
					Identifier("a"),
					BinaryExpr(
						"*",
						UnaryExpr(
							"-",
							Identifier("b"),
						),
						Identifier("c"),
					),
				),
				Identifier("d"),
			),
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(st *testing.T) {
			c.input += "\n"
			defer assertEOF(t, "", false)

			p := stringParser(t, c.input)
			c.assert(st, p.parseExpr())
		})
	}
}

func TestParseExpr_NonAssocOp(t *testing.T) {
	t.Run("followed by other non-assoc op", func(t *testing.T) {
		input := `a == b == c`
		defer assertEOF(t, input, true)

		p := stringParser(t, input)
		defer p.sess.Emit()
		p.parseExpr()
	})

	t.Run("followed by other assoc op", func(t *testing.T) {
		input := `a == b + c`
		defer assertEOF(t, input, false)

		p := stringParser(t, input)
		expr := p.parseExpr()

		BinaryExpr(
			"==",
			Identifier("a"),
			BinaryExpr(
				"+",
				Identifier("b"),
				Identifier("c"),
			),
		)(t, expr)
	})

	t.Run("followed by other non-assoc op with different precedence", func(t *testing.T) {
		input := `a == b :> c`
		defer assertEOF(t, input, false)

		p := stringParser(t, input)
		expr := p.parseExpr()

		BinaryExpr(
			"==",
			Identifier("a"),
			BinaryExpr(
				":>",
				Identifier("b"),
				Identifier("c"),
			),
		)(t, expr)
	})
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
	} else if eof {
		require.FailNow(t, "expected error", input)
	}
}

func stringParser(t *testing.T, str string) *parser {
	scanner := scanner.New("test", strings.NewReader(str))
	scanner.Run()
	loader := source.NewMemLoader()
	loader.Add("test", str)
	cm := source.NewCodeMap(loader)
	require.NoError(t, cm.Add("test"))
	d := diagnostic.NewDiagnoser(cm, diagnostic.Stderr(true, true))

	opTable := operator.NewTable()
	opTable.Add("<|", "", operator.Right, 0)
	opTable.Add("+", "", operator.Left, 6)
	opTable.Add("-", "", operator.Left, 6)
	opTable.Add("*", "", operator.Left, 7)
	opTable.Add("==", "", operator.NonAssoc, 4)
	opTable.Add(":>", "", operator.NonAssoc, 5)

	sess := NewSession(d, cm, opTable)
	var p = newParser(sess)
	p.init("test", scanner, FullParse)
	return p
}