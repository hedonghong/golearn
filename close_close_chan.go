package main

func main()  {
	ch := make(chan int)
	close(ch)
	close(ch)
	//panic: close of closed channel
}
