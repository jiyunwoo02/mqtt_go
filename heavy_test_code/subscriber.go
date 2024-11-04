// 발행자가 exit 입력시 구독자도 종료되어야 하는데 종료되지 않음
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 각 구독자의 결과를 저장하는 구조체
type SubscriberResult struct {
	ID           int  // 구독자 클라이언트 아이디
	ReceivedMQTT int  // MQTT를 통해 받은 메시지 수
	Expected     int  // 기대 메시지 수
	IsSuccessful bool // 수신 성공 여부
}

// MQTT 메시지 수신 함수
func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup, results chan<- SubscriberResult, stopChan chan struct{}, closeOnce *sync.Once) {
	defer wg.Done()

	receivedCount := 0
	expectedCount := 0

	// 발행 횟수 확인 주제 (topic/count) 구독
	countTopic := topic + "/count"
	client.Subscribe(countTopic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		countPayload := string(msg.Payload())
		if count, err := strconv.Atoi(countPayload); err == nil {
			expectedCount += count
			fmt.Printf("-> Subscriber %d: Set expected count to %d\n", id, expectedCount)
		}
	})

	// 발행자 종료 확인 주제 (topic/exit) 구독
	exitTopic := topic + "/exit"
	client.Subscribe(exitTopic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		if string(msg.Payload()) == "exit" {
			fmt.Printf("-> Received exit message. Shutting down subscriber %d.\n", id)
			closeOnce.Do(func() { // 한 번만 stopChan을 닫도록 보장
				close(stopChan)
			})
		}
	})

	// MQTT 주제 구독 및 메시지 수신 핸들러 설정 (QoS=2)
	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		select {
		case <-stopChan:
			return
		default:
			payload := string(msg.Payload())
			fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)
			receivedCount++

			results <- SubscriberResult{
				ID:           id,
				ReceivedMQTT: receivedCount,
				Expected:     expectedCount,
				IsSuccessful: receivedCount == expectedCount,
			}
		}
	})
	<-stopChan
}

// 소켓 메시지를 수신하고 출력하는 함수
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
	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
	id := flag.String("id", "subscriber1", "The id of the subscriber")
	topic := flag.String("tpc", "test/topic", "MQTT topic")
	sn := flag.Int("sn", 1, "Number of subscribers")
	port := flag.String("p", "", "Port to connect for publisher connections")

	flag.Parse()

	var wg sync.WaitGroup
	var conn net.Conn
	var err error
	var closeOnce sync.Once
	results := make(chan SubscriberResult, *sn)
	stopChan := make(chan struct{})

	if *port != "" {
		conn, err = net.Dial("tcp", "localhost:"+*port)
		if err != nil { // 발행자가 소켓을 닫아버렸을 때 수신자가 연결을 시도하지 않도록
			fmt.Println("-- Publisher not using port. Switching to MQTT only.")
		} else {
			fmt.Println("-- Connected to publisher socket.")
			wg.Add(1)
			go receiveFromSocket(conn, &wg)
		}
	} else {
		fmt.Println("-- No port provided on Subscriber. Only MQTT subscription will occur.")
	}

	for i := 1; i <= *sn; i++ {
		wg.Add(1)

		clientID := fmt.Sprintf("%s_%d", *id, i)
		opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(clientID)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Fatalf("-- Failed to connect to broker: %v", token.Error())
		}
		defer client.Disconnect(250)

		go subscribeToMQTT(client, *topic, i, &wg, results, stopChan, &closeOnce)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	totalMessages := 0
	totalReceived := 0
	successfulCount := 0
	unsuccessfulSubscribers := []int{}

	for result := range results {
		totalMessages += result.Expected
		totalReceived += result.ReceivedMQTT

		if result.ReceivedMQTT == result.Expected {
			successfulCount++
			fmt.Printf("=> Subscriber %d: Successfully received %d/%d messages.\n", result.ID, result.ReceivedMQTT, result.Expected)
		} else {
			unsuccessfulSubscribers = append(unsuccessfulSubscribers, result.ID)
		}
	}

	// 요약 결과 출력
	fmt.Printf("\n모든 구독자(%d명) 중 %d명이 메시지를 정상적으로 수신했습니다.\n", *sn, successfulCount)
	fmt.Printf("정상 수신 구독자 수: %d\n", successfulCount)
	fmt.Printf("비정상 수신 구독자 수: %d\n", len(unsuccessfulSubscribers))
	for _, id := range unsuccessfulSubscribers {
		fmt.Printf("- Subscriber%d\n", id)
	}

	fmt.Printf("총 구독자 수: %d\n", *sn)
	fmt.Printf("성공적으로 수신한 메시지: %d/%d (%.1f%%)\n", totalReceived, totalMessages, (float64(totalReceived)/float64(totalMessages))*100)

	fmt.Println(">> All subscribers shutting down.")
}
