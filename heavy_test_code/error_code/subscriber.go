package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 각 구독자의 결과를 저장하는 구조체
type SubscriberResult struct {
	ID           int
	ReceivedMQTT int // MQTT를 통해 받은 메시지 수
	ReceivedSock int // 소켓을 통해 받은 메시지 수
	Expected     int // 기대 메시지 수
	IsSuccessful bool
}

// MQTT 메시지 수신 함수
func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup, results chan<- SubscriberResult) {
	received := 0
	expected := 0

	// 구독자 ID에 따라 대기 시간 설정 -- (구독자 1은 3초, 구독자 2는 2초, 구독자 3은 1초)
	// -- 구독자들이 MQTT 서버에 메시지 수신을 시작하기 전에 기다리게 하는 역할
	// => 구독자 2가 2초 후에 메시지를 수신하려고 준비가 되었을 때, 이미 구독자 3이 메시지를 먼저 수신하고 처리했을 수 있음
	// - MQTT는 기본적으로 한 번 발행된 메시지는 구독자가 구독을 시작한 시점 이후에만 수신할 수 있다.
	// 이 경우, 메시지가 구독자 3에게만 전달되었고, 이후 구독자 2와 1이 구독을 시작했기 때문에 수신하지 못했을 수 있다.
	sleepDuration := time.Duration(4-id) * time.Second
	fmt.Printf("Subscriber %d: Waiting for %v \n", id, sleepDuration)
	time.Sleep(sleepDuration)

	// 메시지 수신 시 호출되는 콜백 함수: 메시지 핸들러
	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		payload := string(msg.Payload())
		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)
		if expected == 0 && strings.Contains(payload, "n=") {
			parts := strings.Split(payload, ", ")
			for _, part := range parts {
				if strings.HasPrefix(part, "n=") {
					nValue := strings.TrimPrefix(part, "n=")
					expected, _ = strconv.Atoi(nValue)
					fmt.Printf("=> Subscriber %d: Extracted n value: %d\n", id, expected)
					break
				}
			}
		} else {
			received++
		}
	})

	//time.Sleep(10 * time.Second)

	results <- SubscriberResult{
		ID:           id,
		ReceivedMQTT: received,
		Expected:     expected,
		IsSuccessful: received == expected,
	}

	wg.Done()
	fmt.Println(time.Now())
	fmt.Print("Done 호출\n")
}

// 소켓 메시지 수신 함수
func receiveFromSocket(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	if conn != nil {
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			message := scanner.Text()
			fmt.Printf(">> Received from TCP Socket: %s\n", message)
		}
	}
}

func main() {
	t := time.Now()
	fmt.Println(t)
	fmt.Print("\n")

	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
	id := flag.String("id", "subscriber1", "The id of the subscriber")
	topic := flag.String("tpc", "test/topic", "MQTT topic")
	sn := flag.Int("sn", 1, "Number of subscribers")                          // sn 플래그 추가 (sn : 구독자의 수)
	port := flag.String("p", "", "Port to connect for publisher connections") // p 플래그 추가 (p : 서버 포트, TCP 소켓 클라이언트 역할)
	flag.Parse()

	var wg sync.WaitGroup
	results := make(chan SubscriberResult, *sn) // 구독자 수만큼 결과 저장할 채널

	var conn net.Conn
	if *port == "" {
		fmt.Printf("No port provided. No socket server connection will be established.\n")
	}
	if *port != "" {
		var err error
		conn, err = net.Dial("tcp", "localhost:"+*port) // The Dial function connects to a server:
		if err != nil {
			log.Fatalf("Failed to connect to publisher via socket: %v\n", err)
		}
		defer conn.Close() // Close closes the connection.
		fmt.Println("Connected to publisher via socket.\n")
		wg.Add(1)
		go receiveFromSocket(conn, &wg)
	}

	for i := 1; i <= *sn; i++ {
		wg.Add(1)
		fmt.Print("Add 호출\n")
		clientID := fmt.Sprintf("%s_%d", *id, i)
		opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(clientID)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Fatalf("-- Failed to connect to broker: %v", token.Error())
		}
		defer client.Disconnect(250)
		go subscribeToMQTT(client, *topic, i, &wg, results)
	}

	go func() {
		wg.Wait()
		fmt.Print("Wait 종료\n")
		close(results)
		fmt.Print("channel close\n")
	}()

	fmt.Print("for문 시작\n")
	for result := range results {
		fmt.Print("for문 순회\n")
		if result.IsSuccessful {
			fmt.Printf("Subscriber %d: Successfully received %d/%d messages.\n", result.ID, result.ReceivedMQTT, result.Expected)
		} else {
			fmt.Printf("Subscriber %d: Failed to receive all messages (%d/%d).\n", result.ID, result.ReceivedMQTT, result.Expected)
		}
	}
	fmt.Print("for문 종료\n")
	fmt.Println(time.Now())
	fmt.Printf("실행 시간: %s\n", time.Since(t).Seconds())
}
