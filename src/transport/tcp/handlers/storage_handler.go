package handlers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
  "time"
	"github.com/minikeyvalue/src/config"
	"github.com/minikeyvalue/src/utils/constants"
	"go.uber.org/zap"
)

type storageInterface interface {
	Set(key string, value string, ttl time.Date) error
	Get(key string) (string, error)
	Del(key string) error
}

type storageHandler struct {
	logger  *zap.Logger
	conn    net.Conn
	storage storageInterface
	cfg     *config.Config
}

func NewStorageHandler(logger *zap.Logger, conn net.Conn,
	storage storageInterface, cfg *config.Config) *storageHandler {
	return &storageHandler{
		logger:  logger,
		conn:    conn,
		storage: storage,
		cfg:     cfg,
	}
}

func (s *storageHandler) HandleClient() {
	defer s.conn.Close()

	isAuth := false

	if err := s.formatAndSendBytes(constants.EnterUserName); err != nil {
		s.logger.Error("Failed to send username request", zap.Error(err))
		return
	}

	username, err := s.readUserMessage()
	if err != nil {
		s.logger.Error("Failed to read username", zap.Error(err))
		s.formatAndSendBytes("Failed to read message, please, try again.")
		return
	}

	if username != s.cfg.UserName {
		s.formatAndSendBytes(constants.IncorrectUserData)
		return
	}

	if err := s.formatAndSendBytes(constants.EnterUserPassword); err != nil {
		s.logger.Error("Failed to send password request", zap.Error(err))
		return
	}

	password, err := s.readUserMessage()
	if err != nil {
		s.logger.Error("Failed to read password", zap.Error(err))
		s.formatAndSendBytes("Failed to read message, please, try again.")
		return
	}

	if password != s.cfg.UserPassword {
		s.formatAndSendBytes(constants.IncorrectUserData)
		return
	}

	isAuth = true
	if err := s.formatAndSendBytes(constants.SuccessfulLogin); err != nil {
		s.logger.Error("Failed to send login success message", zap.Error(err))
		return
	}

	for {
		message, err := s.readUserMessage()
		if err != nil {
			s.logger.Error("Failed to read user message", zap.Error(err))
			s.formatAndSendBytes("Failed to read message, please, try again.")
			return
		}

		if s.cfg.Logging {
			s.logger.Info("Request", zap.String("message", message))
		}

		if !isAuth {
			s.formatAndSendBytes(constants.IncorrectUserData)
			continue
		}

		if err := s.distributeCommands(message); err != nil {
			errMsg := fmt.Sprintf("Error: %s", err.Error())
			s.logger.Error("Failed processing request", zap.Error(err))
			s.formatAndSendBytes(errMsg)
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

		if err := s.formatAndSendBytes(data); err != nil {
			return fmt.Errorf("distributeCommands: failed send data to user: %w", err)
		}
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
	if err := s.formatAndSendBytes("PONG"); err != nil {
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
  defaultTimeTTL := time.Minute * 10
  defaultTTL := time.Now().UTC().Add(defaultTimeTTL)
	if err := s.storage.Set(key, value, defaultTTL); err != nil {
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

func (s *storageHandler) readUserMessage() (string, error) {
	reader := bufio.NewReader(s.conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("readUserMessage: failed to read user message: %w", err)
	}

	return strings.TrimSpace(message), nil
}

func (s *storageHandler) formatAndSendBytes(msg string) error {
	if _, err := s.conn.Write([]byte(msg + "\n")); err != nil {
		return fmt.Errorf("formatAndSendBytes: failed send bytes to user: %w", err)
	}

	return nil
}
