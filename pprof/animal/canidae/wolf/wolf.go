package wolf

import (
	"log"
)

type Wolf struct {
	buffer []byte
}

func (w *Wolf) Name() string {
	return "wolf"
}

func (w *Wolf) Live() {
	w.Eat()
	w.Drink()
	w.Shit()
	w.Pee()
	w.Run()
	w.Howl()
}

func (w *Wolf) Eat() {
	log.Println(w.Name(), "eat")
}

func (w *Wolf) Drink() {
	log.Println(w.Name(), "drink")
	//TODO 下面代码会引起协程过多，注意没意义的协程
	//for i := 0; i < 10; i++ {
	//	go func() {
	//		time.Sleep(30 * time.Second)
	//	}()
	//}
}

func (w *Wolf) Shit() {
	log.Println(w.Name(), "shit")
}

func (w *Wolf) Pee() {
	log.Println(w.Name(), "pee")
}

func (w *Wolf) Run() {
	log.Println(w.Name(), "run")
}

func (w *Wolf) Howl() {
	log.Println(w.Name(), "howl")
	//TODO 下面会引起锁竞争
	//m := &sync.Mutex{}
	//m.Lock()
	//go func() {
	//	time.Sleep(time.Second)
	//	m.Unlock()
	//}()
	//m.Lock()
	/*
	可以看到，这个锁由主协程 Lock，并启动子协程去 Unlock，
	主协程会阻塞在第二次 Lock 这儿等待子协程完成任务，
	但由于子协程足足睡眠了一秒，导致主协程等待这个锁释放足足等了一秒钟。
	虽然这可能是实际的业务需要，逻辑上说得通，并不一定真的是性能瓶颈，
	但既然它出现在我写的“炸弹”里，就肯定不是什么“业务需要”啦。
	 */
}
