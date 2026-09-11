package driver

import "fmt"

var registry = map[string]func() AnyDriver{}

func Register(name string, factory func() AnyDriver) {
	registry[name] = factory
}

func New(name string) (AnyDriver, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", name)
	}
	return f(), nil
}

func Registered() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
