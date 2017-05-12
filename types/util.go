package types

import (
	"fmt"
	"strings"
)

type strSet struct {
	elems []string
	index map[string]struct{}
}

func newStrSet() *strSet {
	return &strSet{index: make(map[string]struct{})}
}

func (s *strSet) add(str string) {
	if _, ok := s.index[str]; !ok {
		s.elems = append(s.elems, str)
		s.index[str] = struct{}{}
	}
}

func (s *strSet) contains(str string) bool {
	_, ok := s.index[str]
	return ok
}

type nameState struct {
	normals     int
	numbers     int
	comparables int
	appendables int
	compappends int
	taken       *strSet
}

func newNameState(taken *strSet) *nameState {
	return &nameState{taken: taken}
}

func (s *nameState) newVar(name string) string {
	for {
		var v string
		switch {
		case strings.HasPrefix(name, number):
			v = s.nextCategoryVar(&s.numbers, number)
		case strings.HasPrefix(name, comparable):
			v = s.nextCategoryVar(&s.comparables, comparable)
		case strings.HasPrefix(name, appendable):
			v = s.nextCategoryVar(&s.appendables, appendable)
		case strings.HasPrefix(name, compappend):
			v = s.nextCategoryVar(&s.compappends, compappend)
		default:
			v = s.nextVar()
		}

		if !s.taken.contains(v) {
			s.taken.add(v)
			return v
		}
	}
}

const (
	startLetter = 'a'
	numLetters  = 'z' - 'a' + 1
)

// nextVar returns the next normal variable.
func (s *nameState) nextVar() string {
	n := s.normals / numLetters
	letter := startLetter + (s.normals % numLetters)
	s.normals++
	if n == 0 {
		return string(letter)
	}

	return fmt.Sprintf("%c%d", letter, n)
}

// nextCategoryVar returns the next variable in a category.
// It receives the pointer to the category counter that will be incremented.
func (s *nameState) nextCategoryVar(category *int, name string) string {
	(*category)++
	return fmt.Sprintf("%s%d", name, *category)
}
