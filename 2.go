package main


type Person2 struct {
	age int
}


func main() {
	var b = Person2{111}
	var a = &b
	println(a)
}

