package main

type set struct {
	m map[string]struct{}
}

func newSet() *set {
	return &set{
		m: make(map[string]struct{}),
	}
}

func (s *set) Add(elem string) {
	s.m[elem] = struct{}{}
}

func (s *set) Export() []string {
	exp := make([]string, 0, len(s.m))
	for elem := range s.m {
		exp = append(exp, elem)
	}
	return exp
}

func (s *set) Contains(elem string) bool {
	_, ok := s.m[elem]
	return ok
}

func (s *set) Extend(elems []string) {
	for _, imp := range elems {
		s.Add(imp)
	}
}
