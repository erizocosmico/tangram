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

type writerEmitter struct {
	w        io.Writer
	warnings bool
	colors   bool
}

func (e *writerEmitter) Emit(file string, ds []Diagnostic) error {
	if !e.warnings {
		var hasErrors bool
		for _, d := range ds {
			if d.Severity() != Warn {
				hasErrors = true
				break
			}
		}

		if !hasErrors {
			return nil
		}
	}

	if err := e.print("I found problems at file: %s\n\n", file); err != nil {
		return err
	}

	for _, d := range ds {
		if err := e.emitDiagnostic(file, d); err != nil {
			return err
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

	return e.print("at %s:%d:%d\n\n", file, d.Line(), d.Column())
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
	var (
		buf            bytes.Buffer
		lines          = d.Lines()
		line           = d.Line()
		maxLine        = lines[len(lines)-1].Num
		lastLineDigits = int64(len(fmt.Sprint(maxLine)))
		lineFormat     = "\n" + `%-` + fmt.Sprint(lastLineDigits) + "d| "
	)

	buf.WriteRune('\n')
	for _, l := range lines {
		buf.WriteString(fmt.Sprintf(lineFormat, l.Num))
		buf.WriteString(l.Content)
		if l.Num == line {
			buf.WriteRune('\n')
			for i := int64(0); i < d.Column()+2-1+lastLineDigits; i++ {
				buf.WriteRune(' ')
			}

			if e.colors {
				buf.WriteString(d.Severity().Color()("^"))
			} else {
				buf.WriteRune('^')
			}
		}
		buf.WriteRune('\n')
	}

	return e.print(buf.String())
}

// Stderr creates a new emitter that will report to stderr all diagnostics.
func Stderr(warnings, colors bool) Emitter {
	return &writerEmitter{os.Stderr, warnings, colors}
}
