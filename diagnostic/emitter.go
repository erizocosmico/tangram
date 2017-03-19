package diagnostic

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Emitter emits diagnostics to the user.
type Emitter interface {
	// Emit emits the given diagnostics for the given file.
	Emit(string, []Diagnostic) error
}

// Errors is an emitter that emits Go errors with the diagnostics.
func Errors(warnings bool) Emitter {
	return &errorEmitter{warnings}
}

type errorEmitter struct {
	warnings bool
}

func (e *errorEmitter) Emit(file string, ds []Diagnostic) error {
	var buf bytes.Buffer
	emitter := writerEmitter{&buf, e.warnings, false}
	if err := emitter.Emit(file, ds); err != nil {
		return err
	}

	return fmt.Errorf("problems found at file: %s\n\n%s", file, buf.String())
}

type writerEmitter struct {
	w        io.Writer
	warnings bool
	colors   bool
}

func hasErrors(ds []Diagnostic) bool {
	for _, d := range ds {
		if d.Severity() != Warn {
			return true
		}
	}
	return false
}

func (e *writerEmitter) Emit(file string, ds []Diagnostic) error {
	if !e.warnings && !hasErrors(ds) {
		return nil
	}

	if err := e.print("I found problems at file: %s\n\n", file); err != nil {
		return err
	}

	for _, d := range ds {
		if e.warnings || d.Severity() != Warn {
			if err := e.emitDiagnostic(file, d); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *writerEmitter) emitDiagnostic(file string, d Diagnostic) error {
	if err := e.printSeverity(d); err != nil {
		return err
	}

	if err := e.print(d.Msg()); err != nil {
		return err
	}

	if d.HasRegion() {
		if err := e.printRegion(d); err != nil {
			return err
		}
	}

	return e.print("\nat %s:%d:%d\n\n", file, d.Line(), d.Column())
}

func (e *writerEmitter) print(msg string, args ...interface{}) error {
	_, err := fmt.Fprintf(e.w, msg, args...)
	return err
}

func (e *writerEmitter) printSeverity(d Diagnostic) error {
	s := d.Severity().String()
	if e.colors {
		s = d.Severity().Color()(s)
	}
	return e.print("%s: ", s)
}

func (e *writerEmitter) printRegion(d Diagnostic) error {
	var buf bytes.Buffer
	lines := d.Lines()
	if len(lines) == 0 {
		return nil
	}

	line := int(d.Line())
	startLine := int(d.StartLine())
	maxLine := startLine + len(lines) - 1
	lastLineDigits := int64(len(fmt.Sprint(maxLine)))
	lineFormat := "\n" + `%-` + fmt.Sprint(lastLineDigits) + "d | "

	buf.WriteRune('\n')
	for i, l := range lines {
		buf.WriteString(fmt.Sprintf(lineFormat, startLine+i))
		buf.WriteString(l)
		if startLine+i == line {
			buf.WriteRune('\n')
			for j := int64(0); j < d.Column()+3-1+lastLineDigits; j++ {
				if e.colors {
					buf.WriteString(d.Severity().Color()("-"))
				} else {
					buf.WriteRune('-')
				}
			}

			if e.colors {
				buf.WriteString(d.Severity().Color()("^"))
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
