package reactive

import "reflect"

type dependencySource interface {
	addDependent(dependentNode)
	removeDependent(dependentNode)
}

type dependentNode interface {
	markDirty()
}

type dependencyCollector interface {
	trackDependency(dependencySource)
}

type sourceNode struct {
	dependents map[dependentNode]struct{}
}

func (s *sourceNode) addDependent(node dependentNode) {
	if s.dependents == nil {
		s.dependents = map[dependentNode]struct{}{}
	}
	s.dependents[node] = struct{}{}
}

func (s *sourceNode) removeDependent(node dependentNode) {
	if s.dependents == nil {
		return
	}
	delete(s.dependents, node)
}

func (s *sourceNode) notifyDependents() {
	if len(s.dependents) == 0 {
		return
	}
	dependents := make([]dependentNode, 0, len(s.dependents))
	for dependent := range s.dependents {
		dependents = append(dependents, dependent)
	}
	for _, dependent := range dependents {
		dependent.markDirty()
	}
}

type dependencySet struct {
	deps []dependencySource
	seen map[dependencySource]struct{}
}

func (d *dependencySet) clear(owner dependentNode) {
	for _, dep := range d.deps {
		dep.removeDependent(owner)
	}
	d.deps = nil
	d.seen = nil
}

func (d *dependencySet) track(owner dependentNode, source dependencySource) {
	if d.seen == nil {
		d.seen = map[dependencySource]struct{}{}
	}
	if _, ok := d.seen[source]; ok {
		return
	}
	d.seen[source] = struct{}{}
	d.deps = append(d.deps, source)
	source.addDependent(owner)
}

func defaultEqual[T any](a, b T) bool {
	return reflect.DeepEqual(a, b)
}
