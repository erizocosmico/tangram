package report

import (
	"errors"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/source"
	"github.com/elm-tangram/tangram/token"
	"github.com/fatih/color"
)

type ReportType byte

const (
	OtherError ReportType = iota
	SyntaxError
	NameError
	TypeError
	Info
	Warning
)

func (t ReportType) String() string {
	switch t {
	case SyntaxError:
		return "syntax error"
	case NameError:
		return "name error"
	case TypeError:
		return "type error"
	case Info:
		return "info"
	case Warning:
		return "warning"
	default:
		return "error"
	}
}

func (t ReportType) Color() func(string, ...interface{}) string {
	switch t {
	case Info:
		return color.CyanString
	case Warning:
		return color.YellowString
	default:
		return color.RedString
	}
}

type Report interface {
	Type() ReportType
	Message() string
	Pos() token.Pos
	Region() *Region
}

type BaseReport struct {
	typ    ReportType
	pos    token.Pos
	msg    string
	region *Region
}

func NewBaseReport(typ ReportType, pos token.Pos, msg string, region *Region) BaseReport {
	return BaseReport{typ, pos, msg, region}
}

func (r BaseReport) Type() ReportType { return r.typ }
func (r BaseReport) Message() string  { return r.msg }
func (r BaseReport) Pos() token.Pos   { return r.pos }
func (r BaseReport) Region() *Region  { return r.region }

func AsError(report Report) error {
	return errors.New(report.Message())
}

type Diagnostic struct {
	Type    ReportType
	Message string
	Pos     source.LinePos
	Region  *source.Snippet
}

type Region struct {
	Start token.Pos
	End   token.Pos
}

func RegionFromNode(node ast.Node) *Region {
	return &Region{node.Pos(), node.End()}
}
