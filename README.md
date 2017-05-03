<img src="https://cdn.rawgit.com/elm-tangram/tangram/master/tangram-logo.svg" height="90" alt="Logo" />

[![Build Status](https://travis-ci.org/elm-tangram/tangram.svg?branch=master)](https://travis-ci.org/elm-tangram/tangram) [![Test Coverage](https://codecov.io/github/elm-tangram/tangram/coverage.svg?branch=master)](https://codecov.io/gh/elm-tangram/tangram/branch/master) [![Go Report Card](https://goreportcard.com/badge/github.com/elm-tangram/tangram)](https://goreportcard.com/report/github.com/elm-tangram/tangram)

`tangram` is an effort to bring the Elm language to the backend. It uses [Go](http://golang.org) as the host language, which comes for free with very nice features such as the always-improving garbage collector, static binaries, cross compilation, and a large ecosystem amongst other things.

### Goals

* Be an alternative implementation of the Elm language. The language will always remain exactly the same as Elm. Being the default imports and the libraries in `elm-lang/core` the only thing that will be different.
* Ability to reuse all the Elm code in `elm-lang/core`, `elm-lang/html`, `elm-lang/http` and other important libraries.
* Transparent usage of pure Elm libraries with zero effort.
* Communication between Elm and Go using `port`s and `Native` modules, just like Elm does with JavaScript.

### Why?

Elm is a very simple, pragmatic and well-thought language. It's a perfect fit for the frontend, and `tangram` aims to explore if it will be good on the backend as well.

### Roadmap

- [ ] Type check
- [ ] Generate Go ASTs from Elm ASTs
- [ ] Go interop and `Native` modules
- [ ] Native implementations for `elm-lang/core`
- [ ] Package management
- [ ] Native implementations for `elm-lang/html`
- [ ] Native implementations for `elm-lang/http`

### Contributing

Right now, contributing can be a bit chaotic. Some parts of the code are a mess and filled with TODOs waiting for a refactor that will come when everything is more or less functional. Nonetheless, if you are interested in contributing to the project you are welcomed to do so!

You can take a look at the roadmap and if there's some part you want to work on just open an issue and you'll be guided through the code and such, if you need it.

### License

**tangram** is licensed under the MIT license, see [LICENSE](/LICENSE)
**tangram** is **not** official or related to the `elm-lang` team in any way.
