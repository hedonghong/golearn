package rpcdm

import (
	"encoding/gob"
	"fmt"
	"net"
	"testing"
)

// User  测试用的用户结构体
type User struct {
	Name string
	Age  int
}

// queryUser 模拟查询用户的方法
func queryUser(uid int) (User, error) {
	// Fake data
	user := make(map[int]User)
	user[0] = User{Name: "Foo", Age: 12}
	user[1] = User{Name: "Bar", Age: 13}
	user[2] = User{Name: "Joe", Age: 14}

	// Fake query
	if u, ok := user[uid]; ok {
		return u, nil
	}
	return User{}, fmt.Errorf("user wiht id %d is not exist", uid)
}

func TestClient(t *testing.T) {
	gob.Register(User{}) // gob 编码要注册一下才能编码结构体

	addr := ":4040"

	// 服务端
	srv := NewServer()
	srv.Register("queryUser", queryUser)
	go srv.ListenAndServe(addr)

	// 客户端
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Error(err)
	}
	cli := NewClient(conn)
	var query func(int) (User, error)
	cli.Call("queryUser", &query)

	u, err := query(2)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(u)
}
