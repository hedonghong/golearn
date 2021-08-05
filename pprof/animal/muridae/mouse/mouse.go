package mouse

import (
	"log"

	"golearn/pprof/constant"
)

type Mouse struct {
	buffer [][constant.Mi]byte
}

func (*Mouse) Name() string {
	return "mouse"
}

func (m *Mouse) Live() {
	m.Eat()
	m.Drink()
	m.Shit()
	m.Pee()
	m.Hole()
	m.Steal()
}

func (m *Mouse) Eat() {
	log.Println(m.Name(), "eat")
}

func (m *Mouse) Drink() {
	log.Println(m.Name(), "drink")
}

func (m *Mouse) Shit() {
	log.Println(m.Name(), "shit")
}

func (m *Mouse) Pee() {
	log.Println(m.Name(), "pee")
}

func (m *Mouse) Hole() {
	log.Println(m.Name(), "hole")
}

func (m *Mouse) Steal() {
	log.Println(m.Name(), "steal")
	//TODO 下面代码会造成内存占用很高
	//max := constant.Gi
	//for len(m.buffer) * constant.Mi < max {
	//	m.buffer = append(m.buffer, [constant.Mi]byte{})
	//}
}
