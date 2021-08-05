package canidae

import "golearn/pprof/animal"

type Canidae interface {
	animal.Animal
	Run()
	Howl()
}
