package interpreter

type Environment struct {
	store  map[string]Value
	parent *Environment
}

func NewEnv() *Environment {
	return &Environment{
		store:  make(map[string]Value),
		parent: nil,
	}
}

func NewEnvIn(parent *Environment) *Environment {
	env := NewEnv()
	env.parent = parent
	return env
}

func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.store[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}

func (e *Environment) Has(name string) bool {
	_, ok := e.store[name]
	return ok
}

func (e *Environment) Delete(name string) {
	delete(e.store, name)
}

func (e *Environment) GetAll() map[string]Value {
	return e.store
}