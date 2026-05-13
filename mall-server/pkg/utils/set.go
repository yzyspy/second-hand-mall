package utils

type Set map[string]struct{}

func (s *Set) Add(key string) {
	(*s)[key] = struct{}{}
}

func (s *Set) Contains(key string) bool {
	_, ok := (*s)[key]
	return ok
}

func (s *Set) Len() int {
	return len(*s)
}
