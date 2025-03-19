package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
)

type storageInterface interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Del(key string) error
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

func (s *storageHandler) HandleClient() {
	defer s.conn.Close()

	for {
		reader := bufio.NewReader(s.conn)
		message, err := reader.ReadString('\n')

		s.logger.Info(message)
		if err != nil {
			s.conn.Write([]byte("Failed read message!\n"))
			s.logger.Error("Failed read user message", zap.Error(err))
			return
		}

		message = strings.TrimSpace(message)
		if message == "PING" {
			if err := s.ping(); err != nil {
				errMsg := fmt.Sprintf("Failed create response, error: %s\n",
					err.Error())
				s.logger.Error("Failed create response for PING command", zap.Error(err))
				s.conn.Write([]byte(errMsg))
			}
		}
		if strings.HasPrefix(message, "GET") {
			parts := strings.SplitN(message, " ", 2)
			data, err := s.get(parts[1])
			if err != nil {
				errMsg := fmt.Sprintf("Failed get data from your storage, error: %s\n",
					err.Error())
				s.logger.Error("Failed get data from user storage", zap.Error(err))
				s.conn.Write([]byte(errMsg))
			} else {
				s.conn.Write([]byte(data + "\n"))
			}
		}

		if strings.HasPrefix(message, "SET") {
			parts := strings.SplitN(message, " ", 3)
			if err := s.set(parts[1], parts[2]); err != nil {
				errMsg := fmt.Sprintf("Failed set new data to storage, error: %s\n",
					err.Error())
				s.logger.Error("Failed set data to storage", zap.Error(err))
				s.conn.Write([]byte(errMsg))
			}
		}

		if strings.HasPrefix(message, "DEL") {
			parts := strings.SplitN(message, " ", 2)
			if err := s.del(parts[1]); err != nil {
				errMsg := fmt.Sprintf("Failed delete item from storage: %s\n",
					err.Error())
				s.logger.Error("Failed delete data from storage", zap.Error(err))
				s.conn.Write([]byte(errMsg))
			}
		}
	}

}

func (s *storageHandler) ping() error {
	if _, err := s.conn.Write([]byte("PONG\n")); err != nil {
		return fmt.Errorf("ping: failed send ping message for client: %w", err)
	}
	return nil
}

func (s *storageHandler) get(key string) (string, error) {
	data, err := s.storage.Get(key)
	if err != nil {
		return "", fmt.Errorf("get: failed get data by key %s:%w", key, err)
	}

	return data, nil
}

func (s *storageHandler) set(key string, value string) error {
	if err := s.storage.Set(key, value); err != nil {
		return fmt.Errorf("set: failed set data to storage: %w", err)
	}

	return nil
}

func (s *storageHandler) del(key string) error {
	if err := s.storage.Del(key); err != nil {
		return fmt.Errorf("del: failed delete data from storage: %w", err)
	}
	return nil
}
