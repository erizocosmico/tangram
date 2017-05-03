package report

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/elm-tangram/tangram/source"
)

// Emitter emits reports to the user.
type Emitter interface {
	// Emit emits the given reports for the given file.
	Emit(string, []*Diagnostic) error
}

// Errors is an emitter that emits Go errors with the reports.
func Errors(warnings bool) Emitter {
	return &errorEmitter{warnings}
}

type errorEmitter struct {
	warnings bool
}

func (e *errorEmitter) Emit(file string, diagnostics []*Diagnostic) error {
	var buf bytes.Buffer
	emitter := writerEmitter{&buf, e.warnings, false}
	if err := emitter.Emit(file, diagnostics); err != nil {
		return err
	}

	return fmt.Errorf("problems found at file: %s\n\n%s", file, buf.String())
}

type writerEmitter struct {
	w        io.Writer
	warnings bool
	colors   bool
}

func hasErrors(diagnostics []*Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Type != Warning {
			return true
		}
	}
	return false
}

func (e *writerEmitter) Emit(file string, diagnostics []*Diagnostic) error {
	if !e.warnings && !hasErrors(diagnostics) {
		return nil
	}

	if err := e.print("I found problems at file: %s\n\n", file); err != nil {
		return err
	}

	for _, d := range diagnostics {
		if e.warnings || d.Type != Warning {
			if err := e.emitReport(file, d); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *writerEmitter) emitReport(file string, d *Diagnostic) error {
	if err := e.printType(d.Type); err != nil {
		return err
	}

	if err := e.print(d.Message); err != nil {
		return err
	}

	if d.Region != nil {
		if err := e.printRegion(d.Type, d.Pos, d.Region); err != nil {
			return err
		}
	}

	return e.print("\nat %s:%d:%d\n\n", file, d.Pos.Line, d.Pos.Col)
}

func (e *writerEmitter) print(msg string, args ...interface{}) error {
	_, err := fmt.Fprintf(e.w, msg, args...)
	return err
}

func (e *writerEmitter) printType(typ ReportType) error {
	s := typ.String()
	if e.colors {
		s = typ.Color()(s)
	}
	return e.print("%s: ", s)
}

func (e *writerEmitter) printRegion(typ ReportType, pos source.LinePos, region *source.Snippet) error {
	var buf bytes.Buffer
	if len(region.Lines) == 0 {
		return nil
	}

	maxLine := region.Start + len(region.Lines) - 1
	lastLineDigits := int64(len(fmt.Sprint(maxLine)))
	lineFormat := "\n" + `%-` + fmt.Sprint(lastLineDigits) + "d | "

	buf.WriteRune('\n')
	for i, l := range region.Lines {
		buf.WriteString(fmt.Sprintf(lineFormat, region.Start+i))
		buf.WriteString(l)
		if region.Start+i == pos.Line {
			buf.WriteRune('\n')
			for j := int64(0); j < int64(pos.Col)+3-1+lastLineDigits; j++ {
				if e.colors {
					buf.WriteString(typ.Color()("-"))
				} else {
					buf.WriteRune('-')
				}
			}

			if e.colors {
				buf.WriteString(typ.Color()("^"))
			} else {
				buf.WriteRune('^')
			}
		}
	}
	buf.WriteRune('\n')

	return e.print(buf.String())
}

// Stderr creates a new emitter that will report to stderr all diagnostics.
func Stderr(warnings, colors bool) Emitter {
	return &writerEmitter{os.Stderr, warnings, colors}
}
