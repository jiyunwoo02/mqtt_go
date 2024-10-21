// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"log"
// 	"net"
// 	"sync"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// func receiveFromSocket(conn net.Conn, id int, expected int, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	scanner := bufio.NewScanner(conn)
// 	received := 0

// 	for scanner.Scan() {
// 		fmt.Printf("[Subscriber %d] Received from socket: %s\n", id, scanner.Text())
// 		received++
// 	}

// 	if received == expected {
// 		fmt.Printf("[Subscriber %d] Successfully received all messages via socket.\n", id)
// 	} else {
// 		fmt.Printf("[Subscriber %d] Missed %d messages via socket.\n", id, expected-received)
// 	}
// }

// func subscribeToBroker(client mqtt.Client, topic string, expected int, id int, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	received := 0

// 	client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
// 		fmt.Printf("[Subscriber %d] Received from broker: %s\n", id, msg.Payload())
// 		received++
// 	})

// 	time.Sleep(5 * time.Second) // 메시지 수신 대기

// 	fmt.Printf("[Subscriber %d] Received %d/%d messages from broker.\n", id, received, expected)
// }

// func main() {
// 	// 소켓 연결
// 	conn, err := net.Dial("tcp", "localhost:9090") // dial : connect server
// 	if err != nil {
// 		log.Fatalf("Failed to connect to publisher via socket: %v", err)
// 	}
// 	defer conn.Close()

// 	// MQTT 클라이언트 설정
// 	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("subscriber")
// 	client := mqtt.NewClient(opts)
// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("Failed to connect to broker: %v", token.Error())
// 	}
// 	defer client.Disconnect(250)

// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	// 브로커와 소켓에서 메시지 수신 시작
// 	go receiveFromSocket(conn, 1, 5, &wg) // 기대 메시지 수 5개로 설정
// 	go subscribeToBroker(client, "test/topic", 5, 1, &wg)

// 	wg.Wait()
// 	fmt.Println("All messages received.")
// }
