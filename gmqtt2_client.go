package main

import (
	"fmt"
	"time"

	"github.com/DrmagicE/gmqtt"
)

func main() {
	// 클라이언트 옵션 설정
	clientOptions := &gmqtt.ClientOptions{
		ClientID:     "testClient",
		CleanSession: true,
		Username:     "user",
		Password:     "password",
	}
	// 클라이언트 생성
	client := gmqtt.NewClient(clientOptions)

	// 브로커에 연결
	if err := client.Connect("tcp://localhost:1883"); err != nil {
		panic(err)
	}

	// 연결 확인
	fmt.Println("클라이언트 연결 완료")

	// 메시지 수신 핸들러 설정
	client.Subscribe("test/topic", 1, func(client *gmqtt.Client, msg *gmqtt.Message) {
		fmt.Printf("수신한 메시지: %s\n", string(msg.Payload()))
	})

	// 메시지 발행
	time.Sleep(1 * time.Second)
	client.Publish("test/topic", 1, false, []byte("Hello, MQTT!"))

	// 5초 대기 후 종료
	time.Sleep(5 * time.Second)
	client.Disconnect()
	fmt.Println("클라이언트 종료")
}
