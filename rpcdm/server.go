package rpcdm

import (
	"github.com/siddontang/go-log/log"
	"net"
	"reflect"
)

type Server struct {
	funcs map[string]reflect.Value
}

func NewServer() *Server {
	return &Server{funcs: map[string]reflect.Value{}}
}

//注册处理函数
func (s *Server) Register(name string, function interface{})  {
	if _, ok := s.funcs[name]; ok {
		return
	}
	fVal := reflect.ValueOf(function)
	s.funcs[name] = fVal
}

//监听连接，并且处理连接

func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for  {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn)  {
	srvSession := NewSession(conn)
	data, err := srvSession.Read()
	if err != nil {
		log.Println("session read error:", err)
		return
	}
	requestRPCData, err := decode(data)
	if err != nil {
		log.Println("data decode error:", err)
		return
	}
	f, ok := s.funcs[requestRPCData.Func]
	if !ok {
		log.Printf("unexpected rpc call: function %s not exist", requestRPCData.Func)
		return
	}
	inArgs := make([]reflect.Value, 0, len(requestRPCData.Args))
	for _, arg := range requestRPCData.Args {
		inArgs = append(inArgs, reflect.ValueOf(arg))
	}
	//反射调用方法
	returnValues := f.Call(inArgs)
	outArgs := make([]interface{}, 0, len(returnValues))
	for _, ret := range returnValues {
		outArgs = append(outArgs, ret.Interface())
	}
	replyRPCData := RPCData{
		Func: requestRPCData.Func,
		Args: outArgs,
	}
	replyEncoded, err := encode(replyRPCData)
	if err != nil {
		log.Println("reply encode error:", err)
		return
	}
	err = srvSession.Write(replyEncoded)
	if err != nil {
		log.Println("reply write error:", err)
	}
}

