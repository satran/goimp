package goimp

// Set provides a store of string, without any particular order,
// and no repeated values
type Set struct {
	state map[string]struct{}
}

// NewSet creates a new Set
func NewSet() *Set {
	return &Set{make(map[string]struct{})}
}

// Add adds an element to the set
func (s *Set) Add(elem string) {
	s.state[elem] = struct{}{}
}

// Export retuns a slice of all elements in the Set
func (s *Set) Export() []string {
	exp := make([]string, 0, len(s.state))
	for elem := range s.state {
		exp = append(exp, elem)
	}
	return exp
}

// Contains checks if the Set contains the given element
func (s *Set) Contains(elem string) bool {
	_, ok := s.state[elem]
	return ok
}

// Extend accepts a slice of elements and updates the Set with them
func (s *Set) Extend(elems []string) {
	for _, imp := range elems {
		s.Add(imp)
	}
}