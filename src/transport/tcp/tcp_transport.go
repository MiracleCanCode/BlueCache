package tcp

import (
	"fmt"
	"net"
)

type tcpConn struct {
	Conn net.Listener
}

func NewWithConn(port string) (*tcpConn, error) {
	addr := fmt.Sprintf("localhost:%s", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("NewWithConn: failed create tcp listener: %w", err)
	}
	return &tcpConn{
		Conn: lis,
	}, nil
}

func (t *tcpConn) CloseConn() error {
	if err := t.Conn.Close(); err != nil {
		return fmt.Errorf("CloseConn: failed close tcp connection: %w", err)
	}

	return nil
}
