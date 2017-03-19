package parser

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/source"
)

// ParseMode specifies the type of mode in which the parser will be run.
// ParseMode can be used to only parse certain parts of a file.
type ParseMode int

const (
	// FullParse will parse a module and all the module imported, parsing
	// all the content in all modules.
	FullParse ParseMode = 1 << iota
	// JustModule will parse just the given module, not parsing any of the
	// modules imported.
	JustModule
	// SkipDefinitions will parse only module declaration, imports and fixity
	// declarations.
	SkipDefinitions
	// StderrDiagnostics will send the diagnostics to stderr instead of
	// returning them as an error.
	StderrDiagnostics
	// SkipWarnings will skip the warning diagnostics.
	SkipWarnings
)

// Is reports whether the given flag is present in the current parse mode.
func (pm ParseMode) Is(flag ParseMode) bool {
	return pm&flag > 0
}

// Session represents the current parsing session.
type Session struct {
	*diagnostic.Diagnoser
	*source.CodeMap
	*operator.Table
}

// NewSession creates a new parsing session with a way of diagnosing errors
// and a code map.
func NewSession(
	d *diagnostic.Diagnoser,
	cm *source.CodeMap,
	ops *operator.Table,
) *Session {
	return &Session{d, cm, ops}
}

// Parse will parse the file at the given path and all its imported modules
// with the given mode of parsing.
func Parse(path string, mode ParseMode) (f *ast.File, err error) {
	// TODO: use proper Fs Loader per project
	cm := source.NewCodeMap(source.NewFsLoader("."))
	defer cm.Close()

	var emitter diagnostic.Emitter
	if mode.Is(StderrDiagnostics) {
		emitter = diagnostic.Stderr(!mode.Is(SkipWarnings), true)
	} else {
		emitter = diagnostic.Errors(!mode.Is(SkipWarnings))
	}

	var optable *operator.Table
	if mode.Is(JustModule) {
		optable = operator.BuiltinTable()
	} else {
		optable = operator.NewTable()
	}

	sess := NewSession(diagnostic.NewDiagnoser(cm, emitter), cm, optable)

	p := newParser(sess)
	if err := cm.Add(path); err != nil {
		return nil, err
	}

	source := cm.Source(path)
	s := scanner.New(source.Path, source.Src)

	s.Run()
	p.init(path, s, mode)
	defer catchBailout()
	if !mode.Is(StderrDiagnostics) {
		defer func() {
			err = sess.Emit()
		}()
	} else {
		defer sess.Emit()
	}
	// TODO: follow imports
	f = p.parseFile()
	return
}

// ParseFrom parses the contents of the given reader and returns the
// corresponding AST file. It will only parse itself and not the imported
// modules, even if it's explicitly requested in the ParseMode.
// All parsing errors encountered will be retuned in the error return value,
// even though StderrDiagnostics mode is present in mode.
func ParseFrom(name string, src io.Reader, mode ParseMode) (f *ast.File, err error) {
	loader := source.NewMemLoader()
	var content []byte
	content, err = ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	loader.Add(name, string(content))
	cm := source.NewCodeMap(loader)
	defer cm.Close()

	sess := NewSession(
		diagnostic.NewDiagnoser(cm, diagnostic.Errors(!mode.Is(SkipWarnings))),
		cm,
		operator.BuiltinTable(),
	)

	p := newParser(sess)
	s := scanner.New(name, bytes.NewBuffer(content))
	s.Run()
	p.init(name, s, mode)
	defer catchBailout()
	defer func() {
		err = sess.Emit()
	}()
	f = p.parseFile()
	return

}

// catchBailout catches "bailout", which means parser has exited on purpose
// due to errors during the parsing. If it's not a bailout the error comes from
// somewhere else and is panicked again.
func catchBailout() {
	if r := recover(); r != nil {
		if _, ok := r.(bailout); !ok {
			panic(r)
		}
	}
}
