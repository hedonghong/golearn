package muridae

import "golearn/pprof/animal"

type Muridae interface {
	animal.Animal
	Hole()
	Steal()
}
