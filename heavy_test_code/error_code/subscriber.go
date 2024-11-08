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
	sleepDuration := time.Duration(11-id) * 5 * time.Millisecond // 5보다는 50으로
	fmt.Printf("Subscriber %d: Waiting for %v \n", id, sleepDuration)
	time.Sleep(sleepDuration)

	// 메시지 수신 시 호출되는 콜백 함수: 메시지 핸들러
	fmt.Printf("구독자 %d 구독 시작\n", id)
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

	time.Sleep(200 * time.Second) // 시간 늘려가면서 성능 테스트 - 200초는 ?

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

	// opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	// client := mqtt.NewClient(opts)
	// if token := client.Connect(); token.Wait() && token.Error() != nil {
	// 	log.Fatalf("-- Failed to connect to broker: %v", token.Error())
	// }
	// defer client.Disconnect(250)

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

	// sn값 만큼 각각 독립적인 MQTT 클라이언트 생성
	// -- 각 클라이언트는 브로커와 별도의 연결을 유지, 독립적으로 메시지 수신, 모든 구독자가 동일한 메시지를 동시에 수신
	// -- QoS 수준 높으면 메시지 누락 시 클라이언트 개별적으로 재전송 요청 처리 가능
	// // -- 구독 시작 시점이 독립적
	for i := 1; i <= *sn; i++ {
		wg.Add(1)
		clientID := fmt.Sprintf("%s_%d", *id, i)
		opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(clientID)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Fatalf("-- Failed to connect to broker: %v", token.Error())
		}
		defer client.Disconnect(250)
		go subscribeToMQTT(client, *topic, i, &wg, results)
	}

	//
	// 하나의 클라이언트를 공유하여 여러 구독을 수행
	// -- 모든 구독자가 하나의 연결을 통해 메시지 수신, 단일 연결에서 다수의 구독 처리 가능 but 독립적으로 메시지 수신 못할 수도 있음
	// -- 특정 구독자가 메시지를 먼저 처리하는 상황 발생할 수 있음, QoS 낮으면 일부에게만 전달될 수 있음
	// -- 순차적으로 메시지를 수신하기에 메시지가 여러 번 발행되지 않는 한 일부 구독자는 메시지를 놓칠 가능성 높아짐
	// for i := 1; i <= *sn; i++ {
	// 	wg.Add(1)
	// 	fmt.Print("Add 호출\n")
	// 	go subscribeToMQTT(client, *topic, i, &wg, results)
	// }

	wg.Wait()
	fmt.Print("Wait 종료\n")
	close(results)
	fmt.Print("channel close\n")

	// go func() {
	// 	wg.Wait()
	// 	fmt.Print("Wait 종료\n")
	// 	close(results)
	// 	fmt.Print("channel close\n")
	// }()

	// 고루틴 상관없이 얘네 없으면 메인 고루틴이 바로 끝나버림 -> 프로그램 종료
	// fmt.Print("for문 시작\n")
	// for result := range results {
	// 	fmt.Print("for문 순회\n")
	// 	if result.IsSuccessful {
	// 		fmt.Printf("Subscriber %d: Successfully received %d/%d messages.\n", result.ID, result.ReceivedMQTT, result.Expected)
	// 	} else {
	// 		fmt.Printf("Subscriber %d: Failed to receive all messages (%d/%d).\n", result.ID, result.ReceivedMQTT, result.Expected)
	// 	}
	// }
	// fmt.Print("for문 종료\n")

	fmt.Println(time.Now())
	fmt.Printf("실행 시간: %s\n", time.Since(t).Seconds())
}
