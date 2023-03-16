package library

type ID int

type Named interface {
	GetName() string
}

type Collection[T Named] struct {
	dataset []T
	ids     map[ID]T
	id_rev  map[string]ID
	names   map[string]T
}

func (c *Collection[T]) nextID() ID {
	return ID(len(c.dataset) + 1)
}

func (c *Collection[T]) Insert(data T) ID {
	id := c.nextID()
	name := data.GetName()
	c.dataset = append(c.dataset, data)
	c.ids[id] = data
	c.names[name] = data
	c.id_rev[name] = id
	return id
}

func (c *Collection[T]) All() []T {
	return c.dataset
}

func (c *Collection[T]) ID(data T) ID {
	return c.id_rev[data.GetName()]
}

func (c *Collection[T]) GetByID(id ID) T {
	return c.ids[id]
}

func (c *Collection[T]) GetByName(id string) T {
	return c.names[id]
}

func NewCollection[T Named]() *Collection[T] {
	return &Collection[T]{
		dataset: make([]T, 0),
		ids:     make(map[ID]T),
		names:   make(map[string]T),
		id_rev:  make(map[string]ID),
	}
}
