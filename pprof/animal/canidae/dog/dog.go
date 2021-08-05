package dog

import (
	"log"
)

type Dog struct {
}

func (d *Dog) Name() string {
	return "dog"
}

func (d *Dog) Live() {
	d.Eat()
	d.Drink()
	d.Shit()
	d.Pee()
	d.Run()
	d.Howl()
}

func (d *Dog) Eat() {
	log.Println(d.Name(), "eat")
}

func (d *Dog) Drink() {
	log.Println(d.Name(), "drink")
}

func (d *Dog) Shit() {
	log.Println(d.Name(), "shit")
}

func (d *Dog) Pee() {
	log.Println(d.Name(), "pee")
}

func (d *Dog) Run() {
	log.Println(d.Name(), "run")
	//TODO 下面这段代码会不断申请，并不断释放内存，因为申请了没用
	//_ = make([]byte, 16 * constant.Mi)
	/*
	可尝试一下将 16 * constant.Mi 修改成一个较小的值，重新编译运行，
	会发现并不会引起频繁 GC，原因是在 golang 里，对象是使用堆内存还是栈内存，
	由编译器进行逃逸分析并决定，如果对象不会逃逸，便可在使用栈内存，但总有意外，
	就是对象的尺寸过大时，便不得不使用堆内存。所以这里设置申请 16 MiB 的内存
	就是为了避免编译器直接在栈上分配，如果那样得话就不会涉及到 GC 了
	 */
}

func (d *Dog) Howl() {
	log.Println(d.Name(), "howl")
}
