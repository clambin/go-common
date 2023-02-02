package set

type Set[K comparable] map[K]struct{}

func (s Set[K]) Add(values ...K) {
	for _, value := range values {
		s[value] = struct{}{}
	}
}

func (s Set[K]) Remove(values ...K) {
	for _, value := range values {
		delete(s, value)
	}
}

func (s Set[K]) Has(value K) bool {
	_, found := s[value]
	return found
}

func (s Set[K]) List() []K {
	var list []K
	for k := range s {
		list = append(list, k)
	}
	return list
}

func (s Set[K]) Equals(other Set[K]) bool {
	for key := range s {
		if _, ok := other[key]; !ok {
			return false
		}
	}
	for key := range other {
		if _, ok := s[key]; !ok {
			return false
		}
	}
	return true
}

func (s Set[K]) Copy() Set[K] {
	output := make(map[K]struct{})
	for key := range s {
		output[key] = struct{}{}
	}
	return output
}

func Create[K comparable](values ...K) Set[K] {
	s := make(Set[K])
	for _, value := range values {
		s[value] = struct{}{}
	}
	return s
}

func Union[K comparable](setA, setB Set[K]) Set[K] {
	union := setA.Copy()
	for key := range setB {
		union[key] = struct{}{}
	}
	return union
}

func Intersection[K comparable](setA, setB Set[K]) Set[K] {
	intersection := make(Set[K])
	for key := range setA {
		if setB.Has(key) {
			intersection[key] = struct{}{}
		}
	}
	return intersection
}

func Difference[K comparable](setA, setB Set[K]) Set[K] {
	difference := make(Set[K])
	for key := range setA {
		if _, ok := setB[key]; !ok {
			difference[key] = struct{}{}
		}
	}
	return difference
}
