package main

// Set provides a store of string, without any particular order,
// and no repeated values
type set struct {
	state map[string]struct{}
}

// NewSet creates a new Set
func newSet() *set {
	return &set{make(map[string]struct{})}
}

// Add adds an element to the set
func (s *set) Add(elem string) {
	s.state[elem] = struct{}{}
}

// Export retuns a slice of all elements in the Set
func (s *set) Export() []string {
	exp := make([]string, 0, len(s.state))
	for elem := range s.state {
		exp = append(exp, elem)
	}
	return exp
}

// Contains checks if the Set contains the given element
func (s *set) Contains(elem string) bool {
	_, ok := s.state[elem]
	return ok
}

// Extend accepts a slice of elements and updates the Set with them
func (s *set) Extend(elems ...string) *set {
	for _, imp := range elems {
		s.Add(imp)
	}
	return s
}
