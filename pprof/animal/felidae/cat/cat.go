package cat

import (
	"log"
)

type Cat struct {
}

func (c *Cat) Name() string {
	return "cat"
}

func (c *Cat) Live() {
	c.Eat()
	c.Drink()
	c.Shit()
	c.Pee()
	c.Climb()
	c.Sneak()
}

func (c *Cat) Eat() {
	log.Println(c.Name(), "eat")
}

func (c *Cat) Drink() {
	log.Println(c.Name(), "drink")
}

func (c *Cat) Shit() {
	log.Println(c.Name(), "shit")
}

func (c *Cat) Pee() {
	log.Println(c.Name(), "pee")
	//TODO 下面代码不是睡眠，而是阻塞 一秒后从chan读取数据
	//<-time.After(time.Second)
	/*
	你应该可以看懂，不同于睡眠一秒，这里是从一个 channel 里读数据时，
	发生了阻塞，直到这个 channel 在一秒后才有数据读出，
	这就导致程序阻塞了一秒而非睡眠了一秒。
	这里有个疑点，就是上文中是可以看到有两个阻塞操作的，
	但这里只排查出了一个，我没有找到其准确原因，
	但怀疑另一个阻塞操作是程序监听端口提供 porof 查询时，
	涉及到 IO 操作发生了阻塞，即阻塞在对 HTTP 端口的监听上，
	但我没有进一步考证。
	 */
}

func (c *Cat) Climb() {
	log.Println(c.Name(), "climb")
}

func (c *Cat) Sneak() {
	log.Println(c.Name(), "sneak")
}
