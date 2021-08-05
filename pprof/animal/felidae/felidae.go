package felidae

import "golearn/pprof/animal"

type Felidae interface {
	animal.Animal
	Climb()
	Sneak()
}
