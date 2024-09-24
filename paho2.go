// MQTT 클라이언트를 생성하고, 지정된 토픽을 구독 및 발행하는 과정
package main

import (
	"fmt"
	"os"
	"time"

	// Go용 MQTT 라이브러리인 paho.mqtt.golang
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 메시지를 수신했을 때 호출되는 핸들러 함수
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

// MQTT 서버에 연결됐을 때 호출되는 핸들러 함수
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

// 연결이 끊겼을 때 호출되는 핸들러 함수
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	var broker = "localhost"
	var port = 1883

	opts := mqtt.NewClientOptions()                          // MQTT 클라이언트 옵션 생성
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port)) // 브로커 주소 설정
	opts.SetClientID("go_mqtt_client")                       // 클라이언트 ID 설정
	opts.SetUsername("user")                                 // MQTT 서버 접속을 위한 사용자 이름 설정
	opts.SetPassword("password")                             // MQTT 서버 접속을 위한 비밀번호 설정

	opts.OnConnect = connectHandler            // 연결 성공 시 호출될 핸들러 설정
	opts.OnConnectionLost = connectLostHandler // 연결 실패 시 호출될 핸들러 설정

	client := mqtt.NewClient(opts) // MQTT 클라이언트 생성

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error()) // 연결 오류 출력
		os.Exit(1)
	}

	// Subscribe to a topic
	topic := "test/topic" // 구독할 주제 설정
	qos := 1              // 메시지 품질(Quality of Service) 설정

	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error()) // 구독 오류 출력
		os.Exit(1)
	}
	fmt.Printf("Subscribed to topic: %s\n", topic) // 구독 성공 메시지 출력

	// Publish a message
	messageText := "Hello from Go client"
	token := client.Publish(topic, byte(qos), false, messageText)
	token.Wait()
	fmt.Printf("Published message: %s to topic: %s\n", messageText, topic) // 발행 성공 메시지 출력

	time.Sleep(6 * time.Second) // 메시지 수신 대기

	// Unsubscribe from the topic
	if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error()) // 구독 취소 오류 출력
		os.Exit(1)
	}
	fmt.Printf("Unsubscribed from topic: %s\n", topic) // 구독 취소 성공 메시지 출력

	// Disconnect
	client.Disconnect(250)
	fmt.Println("Sample Client Disconnected") // 연결 종료 메시지 출력
}
