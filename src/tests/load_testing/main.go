package main 

import (
	"net"
	"go.uber.org/zap"
	"fmt"
	"sync"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	addr := fmt.Sprintf("localhost:%s", "3000")
	
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		key := fmt.Sprintf("ilya:%d", i)
		setCommand := fmt.Sprintf("SET %s isthebest", key)
		go func(cmd string) {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)	
			if err != nil {
				logger.Error("Failed connect to tcp server", zap.Error(err))
				return
			}
			defer conn.Close()
			if _, err := conn.Write([]byte(cmd)); err != nil {
				return
			}
		}(setCommand)
	}
	wg.Wait()
}
