package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/minikeyvalue/src/utils/constants"
	"go.uber.org/zap"
)

const MAX_INPUT_SIZE = 4096

type storageInterface interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Del(key string) error
}

type storageHandler struct {
	logger     *zap.Logger
	conn       net.Conn
	storage    storageInterface
	reqLogging bool
}

func NewStorage(logger *zap.Logger, conn net.Conn,
	storage storageInterface, reqLogging bool) *storageHandler {
	return &storageHandler{
		logger:     logger,
		conn:       conn,
		storage:    storage,
		reqLogging: reqLogging,
	}
}

func (s *storageHandler) HandleClient() {
	defer s.conn.Close()

	for {
		reader := bufio.NewReader(s.conn)
		input := make([]byte, MAX_INPUT_SIZE)
		readedMessage, err := reader.Read(input)
		message := string(input[0:readedMessage])
		if err != nil {
			s.conn.Write([]byte("Failed read message!\n"))
			s.logger.Error("Failed read user message", zap.Error(err))
			return
		}

		message = strings.TrimSpace(message)
		if s.reqLogging {
			s.logger.Info("request", zap.String("message", message))
		}
		if err := s.distributeCommands(message); err != nil {
			errMsg := fmt.Sprintf("Error: %s\n", err.Error())
			s.logger.Error("Failed proccesing requests", zap.Error(err))
			s.conn.Write([]byte(errMsg))
		}
	}

}

func (s *storageHandler) distributeCommands(message string) error {
	if message == constants.PING_COMMAND {
		if err := s.ping(); err != nil {
			return fmt.Errorf("Failed create response: %w", err)
		}
	}

	if strings.HasPrefix(message, constants.GET_COMMAND) {
		parts := strings.SplitN(message, " ", 2)
		data, err := s.get(parts[1])
		if err != nil {
			return fmt.Errorf("Failed get data from your storage: %w", err)
		}

		s.conn.Write([]byte(data + "\n"))
	}

	if strings.HasPrefix(message, constants.SET_COMMAND) {
		parts := strings.SplitN(message, " ", 3)
		if err := s.set(parts[1], parts[2]); err != nil {
			return fmt.Errorf("Failed set new data to storage: %w", err)
		}
	}

	if strings.HasPrefix(message, constants.DEL_COMMAND) {
		parts := strings.SplitN(message, " ", 2)
		if err := s.del(parts[1]); err != nil {
			return fmt.Errorf("Failed delete data from storage: %w", err)
		}
	}

	return nil
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
