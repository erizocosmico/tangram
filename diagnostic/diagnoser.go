package diagnostic

import (
	"github.com/erizocosmico/elmo/source"
	"github.com/fatih/color"
)

// Diagnoser is in charge of reporting the diagnostics occurred during any of
// the compilation steps to the user.
type Diagnoser struct {
	cm          *source.CodeMap
	emitter     Emitter
	diagnostics map[string][]Diagnostic
}

// NewDiagnoser creates a new diagnoser.
func NewDiagnoser(cm *source.CodeMap, emitter Emitter) *Diagnoser {
	return &Diagnoser{
		cm:          cm,
		emitter:     emitter,
		diagnostics: make(map[string][]Diagnostic),
	}
}

// IsOK returns true if there are no diagnostics yet.
func (d *Diagnoser) IsOK() bool {
	return len(d.diagnostics) == 0
}

// Emit writes all the diagnostics using the diagnoser's emitter.
func (d *Diagnoser) Emit() error {
	for file, diagnostics := range d.diagnostics {
		if err := d.emitter.Emit(file, diagnostics); err != nil {
			return err
		}
	}
	return nil
}

// Diagnose adds a new diagnostic occurred in some path.
func (d *Diagnoser) Diagnose(path string, dg Diagnostic) {
	d.diagnostics[path] = append(d.diagnostics[path], dg)
}

// Severity of the diagnostic.
type Severity int

const (
	// Warn is a warning.
	Warn Severity = iota
	// Error is an error diagnostic.
	Error
	// Fatal is a fatal diagnostic.
	Fatal
)

func (s Severity) String() string {
	switch s {
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Fatal:
		return "fatal"
	}
	return "unknown"
}

type colorFunc func(string, ...interface{}) string

// Color returns a function to format with the color of the severity.
func (s Severity) Color() colorFunc {
	switch s {
	case Warn:
		return color.YellowString
	case Error:
		return color.RedString
	case Fatal:
		return color.MagentaString
	}
	return color.WhiteString
}
