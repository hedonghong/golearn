package plan9

import (
	"fmt"
	"testing"
)

var a = 999
func get() int//8 个字节
//TEXT ·get(SB), NOSPLIT, $0-8 8 个字节

func TestRefer(t *testing.T) {
	fmt.Println(get())
}
