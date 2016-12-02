# elmo [![Build Status](https://travis-ci.org/mvader/elmo.svg?branch=master)](https://travis-ci.org/mvader/elmo) [![Test Coverage](https://codecov.io/github/mvader/elmo/coverage.svg?branch=master)](https://codecov.io/gh/mvader/elmo/branch/master) [![Go Report Card](https://goreportcard.com/badge/github.com/mvader/elmo)](https://goreportcard.com/report/github.com/mvader/elmo)

**elm(g)o** is a compiler to bring the Elm language to more places other than the frontend. For that, it compiles to the Go language, which enables a interop with it and the usage of its ecosystem.

**NOTE:** For now, this is just a toy project and highly experimental.

### Goals

* Interop between Elm and Go (probably not the other way around)
* Keep as much of the Elm language as possible
* Make as many parts of `elm-lang/core` as possible work out of the box to allow usage of third party elm libraries that only rely on non-frontend-specific `core` parts

### Why?

For fun, mostly. And because I think Elm is a great language and I'd like to use it for more purposes other than frontend. 
The choice of Go as the host language is basically because of its great ecosystem.

### Roadmap

- [x] Scan Elm code
- [ ] Parse scanned Elm code and build AST
  - [x] Parse `module` declaration
  - [x] Parse `import` declarations
  - [x] Parse `type` declarations
  - [x] Parse literals
  - [ ] Parse value declarations
  - [ ] Parse expressions
- [ ] Semantic analysis
- [ ] Generate Go ASTs from Elm ASTs
- [ ] Module management
- [ ] Go interop and `Native` modules
- [ ] Native implementations for `elm-lang/core`
- [ ] Package management

### License

**elmo** is licensed under the MIT license.
**elmo** is **not** official or related to the `elm-lang` team in any way.
