package main

func main()  {
	var ch chan int
	close(ch)
	close(ch)
	//panic: close of nil channel
}

