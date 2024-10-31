package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"

	// "strconv"
	// "strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 각 구독자의 결과를 저장하는 구조체
type SubscriberResult struct {
	ID           int  // 구독자 클라이언트 아이디
	ReceivedMQTT int  // MQTT를 통해 받은 메시지 수
	Expected     int  // 기대 메시지 수 -> 발행자가 발행한 메시지 개수 count? 소켓을 통해 발행되는 n을 사용할 수도 있지만, 소켓을 사용하지 않는 경우도 있기에 고려 필요
	IsSuccessful bool // 수신 성공 여부 (기대 메시지 수와 실제 수신 메시지 수 비교)
}

// MQTT 메시지 수신 함수
func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup) {
	defer wg.Done() // 구독 작업이 끝나면 WaitGroup에서 작업 완료 처리

	// MQTT 주제 구독 및 메시지 수신 핸들러 설정 (QoS=2)
	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		payload := string(msg.Payload())
		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)
		// received++ // 메시지를 수신할 때마다 카운트 증가
		// *expected = received // 예상 수신 메시지 개수 업데이트
	})

	time.Sleep(5 * time.Second) // 메시지 수신 대기
}

// 소켓 메시지를 수신하고 출력하는 함수
func receiveFromSocket(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	if conn != nil {
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			message := scanner.Text()
			fmt.Printf("Received from Socket (once): %s\n", message)
		}
	}
}

func main() {
	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
	id := flag.String("id", "subscriber1", "The id of the subscriber")
	topic := flag.String("tpc", "test/topic", "MQTT topic")
	sn := flag.Int("sn", 1, "Number of subscribers")                          // sn 플래그 추가 (sn : 구독자의 수)
	port := flag.String("p", "", "Port to connect for publisher connections") // p 플래그 추가 (p : 서버 포트, TCP 소켓 클라이언트 역할)

	flag.Parse() // 플래그 파싱

	// MQTT 클라이언트 설정
	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 구독자 생성 및 수신
	var wg sync.WaitGroup
	results := make(chan SubscriberResult, *sn) // 구독자 수만큼 결과 저장할 채널 생성

	// 각 구독자가 별도의 소켓 연결을 사용하도록 변경 -> 메시지 수신 충돌 해결
	// -- TCP 소켓 연결은 클라이언트-서버 간 1:1 통신을 기본으로 한다!
	var conn net.Conn
	var err error

	if *port != "" {
		// Dial을 사용해 서버에 연결
		conn, err = net.Dial("tcp", "localhost:"+*port) // The Dial function connects to a server:
		if err != nil {
			fmt.Println("Publisher not using port. Switching to MQTT only.")
		} else {
			fmt.Println("Connected to publisher socket.")
			wg.Add(1)
			go receiveFromSocket(conn, &wg)
		}
	} else {
		fmt.Println("No port provided on Subscriber. Only MQTT subscription will occur.")
	}

	// 총 기대 메시지 수 = 메시지 발행 횟수 (n) × 구독자 수 (sn)
	// -> 각 구독자는 n번의 메시지를 받는 것이 목표이기 때문에, 각 구독자의 기대 메시지 수는 발행자의 n

	for i := 1; i <= *sn; i++ {
		wg.Add(1)
		go subscribeToMQTT(client, *topic, i, &wg)
	}

	wg.Wait()      // 모든 수신이 완료될 때까지 대기, Wait blocks until the WaitGroup counter is zero.
	close(results) // 모든 작업이 끝난 후 채널 닫기, close : shutting down the channel after the last sent value is received

	// 결과 출력
	result := <-results
	if result.IsSuccessful {
		fmt.Printf("=> Subscriber %d: Successfully received %d/%d messages.\n",
			result.ID, result.ReceivedMQTT, result.Expected)
	} else {
		fmt.Printf("=> Subscriber %d: Failed to receive all messages (%d/%d).\n",
			result.ID, result.ReceivedMQTT, result.Expected)
	}
}
