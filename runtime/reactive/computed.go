package reactive

import "fmt"

type Computed[T any] struct {
	rt          *Runtime
	fn          func() T
	value       T
	initialized bool
	dirty       bool
	evaluating  bool
	deps        dependencySet
	src         sourceNode
}

func (c *Computed[T]) Get() T {
	c.rt.trackDependency(c)
	if c.dirty || !c.initialized {
		c.evaluate()
	}
	return c.value
}

func (c *Computed[T]) evaluate() {
	if c.evaluating {
		panic(fmt.Sprintf("reactive: cyclic computed evaluation for %T", c))
	}
	c.evaluating = true
	c.deps.clear(c)
	defer func() {
		c.evaluating = false
	}()

	c.rt.withCollector(c, func() {
		c.value = c.fn()
	})
	c.initialized = true
	c.dirty = false
}

func (c *Computed[T]) markDirty() {
	if c.dirty {
		return
	}
	c.dirty = true
	c.src.notifyDependents()
}

func (c *Computed[T]) trackDependency(source dependencySource) {
	c.deps.track(c, source)
}

func (c *Computed[T]) addDependent(node dependentNode) {
	c.src.addDependent(node)
}

func (c *Computed[T]) removeDependent(node dependentNode) {
	c.src.removeDependent(node)
}
