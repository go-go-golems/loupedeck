package reactive

type Signal[T any] struct {
	rt    *Runtime
	value T
	equal func(a, b T) bool
	src   sourceNode
}

func (s *Signal[T]) Get() T {
	s.rt.trackDependency(s)
	return s.value
}

func (s *Signal[T]) Set(value T) {
	if s.equal != nil && s.equal(s.value, value) {
		return
	}
	s.value = value
	s.src.notifyDependents()
	s.rt.maybeFlush()
}

func (s *Signal[T]) Update(fn func(T) T) {
	s.Set(fn(s.value))
}

func (s *Signal[T]) addDependent(node dependentNode) {
	s.src.addDependent(node)
}

func (s *Signal[T]) removeDependent(node dependentNode) {
	s.src.removeDependent(node)
}
