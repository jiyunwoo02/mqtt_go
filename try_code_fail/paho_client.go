package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 메시지 수신 핸들러
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

func main() {
	// 클라이언트 옵션 설정
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").       // 브로커 주소
		SetClientID("testClient").               // 클라이언트 ID 설정
		SetUsername("username").                 // 사용자 이름 설정 (필요시)
		SetPassword("password").                 // 비밀번호 설정 (필요시)
		SetKeepAlive(10 * time.Second).          // KeepAlive 간격 설정
		SetDefaultPublishHandler(messageHandler) // 기본 메시지 핸들러 설정

	// 클라이언트 생성
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("브로커에 연결 실패: %v", token.Error())
	}
	fmt.Println("클라이언트가 브로커에 연결됨")

	// test/topic 구독
	if token := client.Subscribe("test/topic", 0, messageHandler); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Println("test/topic 구독 완료")
	}

	// 메시지 발행
	message := "Hello, MQTT!"
	token := client.Publish("test/topic", 0, false, message)
	token.Wait()
	fmt.Printf("메시지 발행: %s\n", message)

	// 3초 동안 대기하여 메시지 수신 대기
	time.Sleep(3 * time.Second)

	// 클라이언트 종료
	client.Disconnect(250)
	fmt.Println("클라이언트 종료됨")
}
