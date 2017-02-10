package parser

import (
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/source"
)

type Session struct {
	*diagnostic.Diagnoser
	*source.CodeMap
}

func NewSession(d *diagnostic.Diagnoser, cm *source.CodeMap) *Session {
	return &Session{d, cm}
}
