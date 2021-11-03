package rpcdm

import (
	"bytes"
	"encoding/gob"
)

//RPC通信的数据格式
type RPCData struct {
	Func string //访问的函数
	Args []interface{} //函数的参数
}

//encode将RPCData编码
func encode(data RPCData) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//decode将数据解码为RPCData
func decode(data []byte) (RPCData, error) {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	var rpcData RPCData
	err := decoder.Decode(&rpcData)
	return rpcData, err
}
