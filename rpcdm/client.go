package rpcdm

import (
	"net"
	"reflect"
)

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Call(name string, funcPtr interface{})  {
	//反射初始化funcPtr函数原型
	fn := reflect.ValueOf(funcPtr).Elem()

	f := func(args []reflect.Value) []reflect.Value {
		inArgs := make([]interface{}, 0, len(args))
		for _, arg := range args {
			inArgs = append(inArgs, arg.Interface())
		}

		cliSession := NewSession(c.conn)

		requestRPCData := RPCData{
			Func: name,
			Args: inArgs,
		}

		requestEncode, err := encode(requestRPCData)
		if err != nil {
			panic(err)
		}
		if err := cliSession.Write(requestEncode); err != nil {
			panic(err)
		}

		response, err := cliSession.Read()
		if err != nil {
			panic(err)
		}
		respRPCData, err := decode(response)
		if err != nil {
			panic(err)
		}
		outArgs := make([]reflect.Value, 0, len(respRPCData.Args))
		for i, arg := range respRPCData.Args {
			if arg == nil {
				outArgs = append(outArgs, reflect.Zero(fn.Type().Out(i)))
			} else {
				outArgs = append(outArgs, reflect.ValueOf(arg))
			}
		}
		return outArgs
	}

	v := reflect.MakeFunc(fn.Type(), f)
	fn.Set(v)
}
