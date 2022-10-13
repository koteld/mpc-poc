package messaging

import (
	"fmt"

	"mpc_poc/helper"

	"github.com/joho/godotenv"
	"github.com/matryer/vice"
	"github.com/matryer/vice/queues/redis"
	goredis "gopkg.in/redis.v3"
)

var transport vice.Transport

const ProtocolMessagesChannel = "protocol:messages"
const InternalMessagesChannel = "internal:messages"
const SessionMessagesChannel = "session:messages"

const LocalAddr = "127.0.0.1:6379"
const LocalPass = ""

func getTransport() vice.Transport {
	if transport == nil {
		_ = godotenv.Load()
		RedisAddr := helper.GetEnv("RedisAddr", LocalAddr)
		RedisPass := helper.GetEnv("RedisPass", LocalPass)
		client := goredis.NewClient(&goredis.Options{
			Network:    "tcp",
			Addr:       RedisAddr,
			Password:   RedisPass,
			DB:         0,
			MaxRetries: 0,
		})
		transport = redis.New(redis.WithClient(client))
	}
	return transport
}

func GetOutputChannel(name string) chan<- []byte {
	if transport == nil {
		getTransport()
	}
	fmt.Printf("GetOutputChannel: %s\n", name)
	return transport.Send(name)
}

func GetInputChannel(name string) <-chan []byte {
	if transport == nil {
		getTransport()
	}
	fmt.Printf("GetInputChannel: %s\n", name)
	return transport.Receive(name)
}
