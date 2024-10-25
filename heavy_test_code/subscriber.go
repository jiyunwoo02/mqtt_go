package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 각 구독자의 결과를 저장하는 구조체
type SubscriberResult struct {
	ID           int
	Received     int
	Expected     int
	IsSuccessful bool
}

func main() {
	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
	address := flag.String("ad", "tcp://localhost:1883", "Address of the broker")
	id := flag.String("id", "subscriber1", "The id of the subscriber")
	topic := flag.String("tp", "test/topic", "MQTT topic")
	sn := flag.Int("sn", 1, "Number of subscribers")
	port := flag.String("p", "", "Port to connect for publisher connections")

	// MQTT 클라이언트 설정
	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 소켓 연결 (포트가 제공된 경우)
	var conn net.Conn
	if *port != "" {
		var err error
		conn, err = net.Dial("tcp", "localhost:"+*port)
		if err != nil {
			log.Fatalf("Failed to connect to publisher via socket: %v", err)
		}
		defer conn.Close()
		fmt.Println("Connected to publisher via socket.")
	}
}
