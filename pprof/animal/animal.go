package animal

import (
	"golearn/pprof/animal/canidae/dog"
	"golearn/pprof/animal/canidae/wolf"
	"golearn/pprof/animal/felidae/cat"
	"golearn/pprof/animal/felidae/tiger"
	"golearn/pprof/animal/muridae/mouse"
)

var (
	AllAnimals = []Animal{
		&dog.Dog{},
		&wolf.Wolf{},

		&cat.Cat{},
		&tiger.Tiger{},

		&mouse.Mouse{},
	}
)

type Animal interface {
	Name() string
	Live()

	Eat()
	Drink()
	Shit()
	Pee()
}
