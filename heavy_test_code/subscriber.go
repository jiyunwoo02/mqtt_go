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
	ID           int // 구독자 클라이언트 아이디
	ReceivedMQTT int // MQTT를 통해 받은 메시지 수
	Expected     int // 기대 메시지 수
}

// MQTT 메시지 수신 함수
func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup, results chan<- SubscriberResult, stopChan chan struct{}) {
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
			fmt.Printf("-> Subscriber %d received exit message. Shutting down.\n", id)
			// close(stopChan) // 모든 구독자에게 종료 신호 전송
		}
	})

	// MQTT 주제 구독 및 메시지 수신 핸들러 설정 (QoS=2)
	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		select {
		case <-stopChan:
			return // stopChan이 닫히면 메시지 수신 중단
		default:
			payload := string(msg.Payload())
			fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)
			receivedCount++

			results <- SubscriberResult{
				ID:           id,
				ReceivedMQTT: receivedCount,
				Expected:     expectedCount,
			}
		}
	})

	<-stopChan // 종료 신호가 올 때까지 대기
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

	results := make(chan SubscriberResult, *sn)
	stopChan := make(chan struct{}) // 종료 신호용 채널

	if *port != "" {
		conn, err = net.Dial("tcp", "localhost:"+*port)
		if err != nil {
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
		go subscribeToMQTT(client, *topic, i, &wg, results, stopChan)
	}

	go func() {
		wg.Wait()      // 모든 고루틴의 작업이 완료될 때까지 대기 -> 조기 종료 문제 방지
		close(results) // 모든 고루틴이 완료된 후에만 그 고루틴이 종료되어 넘어온다
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

	// // 요약 결과 출력 (추후 수정)
	// fmt.Printf("\n모든 구독자(%d명) 중 %d명이 메시지를 정상적으로 수신했습니다.\n", *sn, successfulCount)
	// fmt.Printf("정상 수신 구독자 수: %d\n", successfulCount)
	// fmt.Printf("비정상 수신 구독자 수: %d\n", len(unsuccessfulSubscribers))
	// for _, id := range unsuccessfulSubscribers {
	// 	fmt.Printf("- Subscriber%d\n", id)
	// }

	// fmt.Printf("총 구독자 수: %d\n", *sn)
	// fmt.Printf("성공적으로 수신한 메시지: %d/%d (%.1f%%)\n", totalReceived, totalMessages, (float64(totalReceived)/float64(totalMessages))*100)

	// fmt.Println(">> All subscribers shutting down.")
}

/* 코드 동작 과정
1. 명령줄 인자 파싱 및 초기화
2. result 채널과 stopChan 생성
3. TCP 연결 고루틴 생성 (옵션) : 연결이 성공하면 receiveFromSocket 함수가 고루틴으로 실행, Waitgroup 카운터 1 증가시켜 대기 설정
4. MQTT 구독자 고루틴 생성: sn 값 기준 sn개 고루틴 생성, 카운터 1 증가 - subscribeToMQTT 함수 고루틴 - 종료 신호 전달 때까지 메시지 수신해 results 채널에 전송 - wg.Done() 호출해 고루틴 종료 알림
5. 고루틴 대기 및 채널 종료 : 익명 고루틴 실행, Wait()는 모든 MQTT와 TCP 수신 고루틴 종료 기다림, 모든 구독자가 Done() 호출해 카운터 0되면 Wait() 완료, results 채널 닫음
6. 결과 수신 고루틴 : 메인 고루틴에서 results 채널의 데이터 수신해 결과 처리, 채널 닫힐 때까지 개수 누적
*/

/*
별도의 고루틴에서 `wg.Wait()`와 `close(results)`를 실행할 때 모든 고루틴이 안전하게 종료되는 이유는
`wg.Wait()`가 별도의 고루틴에서 동작하며, 메인 고루틴과 결과 수신 흐름이 분리되기 때문

1. WaitGroup의 동작 원리
   `wg.Wait()`는 모든 고루틴이 완료되어 `wg.Done()`이 호출된 뒤에야 다음으로 넘어갑니다.
   즉, `wg.Wait()`가 호출된 고루틴은 모든 고루틴이 작업을 완료하고 `Done()`을 호출할 때까지 대기하는 특성을 가지고 있습니다.

2. **메인 고루틴의 결과 수신 흐름 유지**
   별도의 고루틴에서 `wg.Wait()`와 `close(results)`를 실행하면, 메인 고루틴은 그동안 `results` 채널에서 결과를 계속 수신하게 됩니다.
   이 구조에서는 `results` 채널이 닫히기 전까지 메인 고루틴은 데이터를 수신하고, `wg.Wait()`가 모든 고루틴이 종료될 때까지 기다립니다.

- 코드 실행 흐름 (별도 고루틴에서 `wg.Wait()`와 `close(results)` 실행)

1. 고루틴 생성 및 작업 시작: 여러 고루틴이 생성되어 각자의 작업을 수행하고 `results` 채널에 데이터를 전송합니다.

2. `wg.Wait()` 대기: 별도의 고루틴에서 `wg.Wait()`가 호출되어 모든 작업이 완료될 때까지 대기합니다.

3. 메인 고루틴의 결과 수신: 메인 고루틴은 `results` 채널에서 데이터를 계속 읽습니다. 채널이 닫히기 전까지는 안전하게 데이터를 읽어올 수 있습니다.

4. 모든 고루틴 완료 후 채널 종료: 모든 고루틴이 `Done()`을 호출하여 `wg.Wait()`가 종료되면, 별도의 고루틴에서 `close(results)`가 호출됩니다.
	- 이 시점에서 모든 고루틴의 작업이 완료되었기 때문에 `results` 채널에 기록될 데이터는 모두 안전하게 기록된 상태

5. 메인 고루틴의 안전한 종료: 메인 고루틴은 `results` 채널이 닫혔음을 감지하고, 마지막 데이터까지 모두 수신한 뒤 종료됩니다.

### 요약
- 별도의 고루틴에서 `wg.Wait()`와 `close(results)`를 실행하면, `results` 채널이 닫히기 전에 모든 고루틴이 종료될 시간을 확보하게 됩니다.
- 메인 고루틴이 `results` 채널에서 안전하게 데이터를 모두 수신할 수 있는 환경이 조성되므로, 일부 고루틴의 결과가 누락되는 문제가 발생하지 않습니다.
*/
