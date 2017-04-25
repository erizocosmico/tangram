package parser

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/package"
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

// ParseResult is the result after a full parse, which is a set of parsed files
// and the order in which they need to be resolved based on their imports.
type ParseResult struct {
	// Resolution contains a set of all the modules in `Modules` ordered from
	// first module that needs to be resolved to the last.
	Resolution []string
	// Modules contains a mapping between module names and ast files. All the
	// modules will be in `Resolution` and if a module is not in here, it won't
	// be in `Resolution`.
	Modules map[string]*ast.Module
}

// Parse will parse the file at the given path and all its imported modules
// with the given mode of parsing.
func Parse(path string, mode ParseMode) (result *ParseResult, err error) {
	pkg, err := pkg.Load(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	cm := source.NewCodeMap(source.NewFsLoader(pkg))
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
	defer catchBailout()
	if !mode.Is(StderrDiagnostics) {
		defer func() {
			err = sess.Emit()
		}()
	} else {
		defer sess.Emit()
	}

	fp := newFullParser(p, pkg, optable, cm)
	result = fp.parse(path)
	return
}

type fullParser struct {
	p       *parser
	pkg     *pkg.Package
	optable *operator.Table
	cm      *source.CodeMap
	g       *pkg.Graph
}

func newFullParser(p *parser, pkg *pkg.Package, optable *operator.Table, cm *source.CodeMap) *fullParser {
	return &fullParser{
		p,
		pkg,
		optable,
		cm,
		nil,
	}
}

func (p *fullParser) parse(path string) *ParseResult {
	// do a first parse to gather all the imports and operator fixities
	p.firstPass(path, make(map[string]struct{}))

	modules, err := p.g.Resolve()
	switch err := err.(type) {
	case *pkg.CircularDependencyError:
		p.error(
			path,
			fmt.Sprintf("I found a circular dependency in your code between these modules:\n- %s\n- %s", err.Modules[0], err.Modules[1]),
		)
	case nil:
	default:
		p.error(
			path,
			fmt.Sprintf("Oops, an unexpected error happened: %s", err.Error()),
		)
	}

	r := &ParseResult{modules, make(map[string]*ast.Module)}
	for _, m := range modules {
		if file := p.completeParse(m); file != nil {
			r.Modules[m] = file
		}
	}

	return r
}

func (p *fullParser) firstPass(path string, visited map[string]struct{}) {
	if err := p.cm.Add(path); err != nil {
		p.error(path, "Oops, unexpected error reading file: %s", err)
		panic(bailout{})
	}
	source := p.cm.Source(path)
	scanner := source.Scanner()

	p.p.init(source.Path, scanner, SkipDefinitions)
	file := parseFile(p.p)

	mod := file.Module.ModuleName()
	// TODO: check module name corresponds to the path
	visited[mod] = struct{}{}
	if p.g == nil {
		p.g = pkg.NewGraph(mod)
	}

	if p.p.mode.Is(JustModule) {
		return
	}

	for _, imp := range file.Imports {
		importMod := imp.ModuleName()
		if _, ok := visited[importMod]; ok {
			continue
		}

		p.g.Add(importMod, mod)
		importPath, err := p.pkg.FindModule(importMod)
		if err != nil {
			p.error(
				path,
				fmt.Sprintf("I could not find module %q in any of the package source directories or any of its dependencies. Maybe you're missing a dependency?", path),
			)
		}
		p.firstPass(importPath, visited)
	}

	for _, d := range file.Decls {
		if fixity, ok := d.(*ast.InfixDecl); ok {
			n, _ := strconv.Atoi(fixity.Precedence.Value)
			p.optable.Add(fixity.Op.Name, mod, fixity.Assoc, uint(n))
		}
	}
}

func (p *fullParser) completeParse(module string) *ast.Module {
	path, err := p.pkg.FindModule(module)
	if err != nil {
		// TODO: fix this, but should be unreachable
		panic(err)
	}

	source := p.cm.Source(path)
	p.p.init(path, source.Scanner(), FullParse)
	return parseFile(p.p)
}

func (p *fullParser) error(path, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	p.p.sess.Diagnose(
		path,
		diagnostic.NewMsgDiagnostic(diagnostic.Error, bytes.NewBuffer([]byte(msg))),
	)
}

// ParseFrom parses the contents of the given reader and returns the
// corresponding AST file. It will only parse itself and not the imported
// modules, even if it's explicitly requested in the ParseMode.
// All parsing errors encountered will be retuned in the error return value,
// even though StderrDiagnostics mode is present in mode.
func ParseFrom(name string, src io.Reader, mode ParseMode) (f *ast.Module, err error) {
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
	f = parseFile(p)
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
