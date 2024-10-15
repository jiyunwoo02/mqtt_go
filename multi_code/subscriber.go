package main

import (
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message on topic [%s]: %s\n", msg.Topic(), msg.Payload())
}

func main() {
	if len(os.Args) < 4 {
		// 인자의 개수가 맞도록, 4 초과하는 경우더라도 정상 동작
		fmt.Println("Usage: go run subscriber.go <broker_address> <client_id> <topic>")
		return
	}

	// 구독자 구동 시 명령행 인자 3개
	// os.Args[0]는 subscriber.go
	brokerAddress := os.Args[1] // 브로커의 주소
	clientID := os.Args[2]      // 구독자 클라이언트의 아이디
	topic := os.Args[3]         // 구독자가 구독할 주제

	subscriberOpts := mqtt.NewClientOptions().
		AddBroker(brokerAddress).
		SetClientID(clientID).
		SetCleanSession(true)

	subscriberClient := mqtt.NewClient(subscriberOpts)
	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자1 브로커 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자1 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

	if token := subscriberClient.Subscribe(topic, 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to subscribe to topic '%s': %v\n", topic, token.Error())
	}
	fmt.Printf("Subscribed to topic '%s'\n", topic)

	select {}
}
