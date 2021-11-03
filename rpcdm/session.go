package rpcdm

import (
	"encoding/binary"
	"io"
	"net"
)

//RPC通话连接
type Session struct {
	conn net.Conn
}

func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

//向Session中写数据
func (s *Session) Write(data []byte) error {
	buf := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	_, err := s.conn.Write(buf)
	return err
}

//从Session中读数据
func (s *Session) Read() ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(s.conn, header); err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(s.conn, data); err != nil {
		return nil, err
	}
	return data, nil
}
