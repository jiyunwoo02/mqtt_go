package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DrmagicE/gmqtt"
)

func main() {
	// 클라이언트 옵션 설정
	opts := &gmqtt.ClientOptions{
		ClientID:     "testClient",
		CleanSession: true,
	}

	// 클라이언트 생성
	client := gmqtt.NewClient(opts)

	// 브로커에 연결
	if err := client.Connect("tcp://localhost:1883"); err != nil {
		log.Fatalf("브로커에 연결 실패: %v", err)
	}

	fmt.Println("클라이언트가 브로커에 연결되었습니다.")

	// 메시지 수신 핸들러 설정
	client.Subscribe("test/topic", 0, func(client *gmqtt.Client, msg *gmqtt.Message) {
		fmt.Printf("수신한 메시지: %s\n", string(msg.Payload))
	})

	// 메시지 발행
	client.Publish(context.Background(), &gmqtt.Message{
		Topic:   "test/topic",
		QOS:     0,
		Payload: []byte("Hello, MQTT!"),
	})

	// 5초 대기 후 종료
	time.Sleep(5 * time.Second)

	// 연결 종료
	client.Close()
	fmt.Println("클라이언트가 종료되었습니다.")
}
