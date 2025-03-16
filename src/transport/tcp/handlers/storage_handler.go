package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
)

type storageInterface interface {
	Add(key string, value any) error
	Get(key string) (any, error)
}

type storageHandler struct {
	logger  *zap.Logger
	conn    net.Conn
	storage storageInterface
}

func NewStorage(logger *zap.Logger, conn net.Conn,
	storage storageInterface) *storageHandler {
	return &storageHandler{
		logger:  logger,
		conn:    conn,
		storage: storage,
	}
}

func (s *storageHandler) HandleClient() error {
	defer s.conn.Close()

	for {
		reader := bufio.NewReader(s.conn)
		message, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("HandleClient: failed read ping message: %w", err)
		}

		message = strings.TrimSpace(message)
		if message == "PING" {
			if err := s.ping(); err != nil {
				return fmt.Errorf("HandleClient: faield send message to client: %w", err)
			}
		}

		if message == "GET" {
			s.get()
		}
	}

}

func (s *storageHandler) ping() error {
	if _, err := s.conn.Write([]byte("PONG\n")); err != nil {
		return fmt.Errorf("ping: failed send ping message for client: %w", err)
	}
	return nil
}

func (s *storageHandler) get() {}
