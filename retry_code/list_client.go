package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mu sync.Mutex         // 동시성 문제 방지 위한 뮤텍스
var publishMsg []string   // 발행자가 발행한 메시지 리스트
var subscribeMsg []string // 구독자가 수신한 메시지 리스트

// 구독자가 메시지를 수신할 때마다 리스트에 메시지 추가 및 메시지 리턴
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	mu.Lock()
	subscribeMsg = append(subscribeMsg, fmt.Sprintf("토픽 [%s] - 메시지 [%s]", msg.Topic(), msg.Payload()))
	mu.Unlock()
	fmt.Printf("수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

// 발행자가 메시지를 발행할 때마다 리스트에 메시지 추가 및 메시지 리턴
func publishMessage(client mqtt.Client, topic, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	mu.Lock()
	publishMsg = append(publishMsg, fmt.Sprintf("토픽 [%s] - 메시지 [%s]", topic, message))
	mu.Unlock()
	fmt.Printf("발행자가 메시지 발행: 토픽 [%s] - 메시지 [%s]\n", topic, message)
}

func main() {
	publisherOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("publisher")

	publisherClient := mqtt.NewClient(publisherOpts)
	if token := publisherClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())

	subscriberOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("subscriber")

	subscriberClient := mqtt.NewClient(subscriberOpts)
	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자 %s 가 브로커 [%s]에 연결됨\n\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

	if token := subscriberClient.Subscribe("test/topic", 1, msgHandler); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자 %s 가 test/topic 구독 완료\n", subscriberOpts.ClientID)
	}

	// 메시지 발행
	publishMessage(publisherClient, "test/topic", "Hello1!")
	publishMessage(publisherClient, "test/topic", "Hello2!")

	// 3초 동안 대기하여 메시지 수신 대기
	time.Sleep(3 * time.Second)

	// 구독자가 수신한 메시지 목록 출력
	mu.Lock()
	fmt.Println("구독자가 수신한 메시지 목록:")
	for _, message := range subscribeMsg {
		fmt.Println(message)
	}
	mu.Unlock()

	// 발행자가 발행한 메시지 목록 출력
	mu.Lock()
	fmt.Println("발행자가 발행한 메시지 목록:")
	for _, message := range publishMsg {
		fmt.Println(message)
	}
	mu.Unlock()

	// 클라이언트 종료
	subscriberClient.Disconnect(250)
	publisherClient.Disconnect(250)

	fmt.Println("클라이언트 종료됨")
}
