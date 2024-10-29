// 얘가 제일 최우선
// 보면 마지막 클라이언트만 2/2 형식으로 반영됨, 나머지는 다 0/0
// 그리고 처음과 마지막 클라이언트만 일부 수신함, 나머지는 아예 수신 안함
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
	ID           int  // 구독자 클라이언트 아이디
	ReceivedMQTT int  // MQTT를 통해 받은 메시지 수
	ReceivedSock int  // 소켓을 통해 받은 메시지 수
	Expected     int  // 기대 메시지 수
	IsSuccessful bool // 수신 성공 여부 (기대 메시지 수와 실제 수신 메시지 수 비교)
}

// MQTT 메시지 수신 함수
func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup, results chan<- SubscriberResult) {
	defer wg.Done() // 구독 작업이 끝나면 WaitGroup에서 작업 완료 처리

	received := 0 // 실제로 수신한 메시지 개수
	expected := 0 // 발행자가 보낸 총 메시지 횟수(n)

	// MQTT 주제 구독 및 메시지 수신 핸들러 설정 (QoS=2)
	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
		payload := string(msg.Payload())
		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)
		received++ // 메시지를 수신할 때마다 카운트 증가

		// 첫 번째로만 n 값을 추출해 기대 메시지 수 설정
		if strings.Contains(payload, "n=") {
			// 발행된 메시지에서 "n=" 이 포함된 부분을 찾기 위해 문자열을 ", " 구분자로 분할
			parts := strings.Split(payload, ", ")

			// 분할된 부분을 순회하며, "n="으로 시작하는 부분을 탐색
			for _, part := range parts {
				if strings.HasPrefix(part, "n=") {
					// "n=" 접두어를 제거하여 숫자 값만 추출
					nValue := strings.TrimPrefix(part, "n=")

					// 추출된 값을 정수로 변환하여 expectedMessages에 저장
					expected, _ = strconv.Atoi(nValue)

					// 추출된 n 값과 해당 구독자 ID를 출력
					// fmt.Printf("-> Subscriber %d: Extracted n value: %d\n", id, expected)

					// n 값을 찾았으면 반복문 종료
					break
				}
			}
		}
	})

	time.Sleep(5 * time.Second) // 메시지 수신 대기

	// 결과 저장
	results <- SubscriberResult{
		ID:           id,
		ReceivedMQTT: received,
		Expected:     expected,
		IsSuccessful: received == expected,
	}
}

// 소켓 메시지 수신 함수
// 발행자가 소켓을 통해 보낸 메시지를 구독자가 수신
func receiveFromSocket(conn net.Conn, id int, wg *sync.WaitGroup, results chan<- SubscriberResult) {
	defer wg.Done() // 이 함수가 끝나면 WaitGroup에서 작업 하나를 완료 처리

	scanner := bufio.NewScanner(conn) // 소켓(데이터 통로) 연결된 conn에서 버퍼(소켓을 통해 전송된 데이터의 임시 저장 공간)로부터 데이터를 한 줄 단위로 읽는다, 데이터를 읽는 즉시 버퍼에서 해당 데이터는 제거됨
	received := 0                     // 수신한 메시지 개수
	expected := 0                     // n 값 추출 후 저장될 변수
	received++                        // 메시지를 수신할 때마다 카운트 증가

	for scanner.Scan() {
		message := scanner.Text()
		fmt.Printf("[Subscriber %d] Received from Socket: %s\n", id, message)

		// 첫 번째로만 n 값을 추출해 기대 메시지 수 설정
		if strings.Contains(message, "n=") {
			parts := strings.Split(message, ", ")
			for _, part := range parts {
				// fmt.Printf("%s\n", part) -> , 을 기준으로 분할된다 : he 발행시 (n=1) -> he#1 과 n=3
				if strings.HasPrefix(part, "n=") {
					nValue := strings.TrimPrefix(part, "n=")
					expected, _ = strconv.Atoi(nValue) // n 값을 추출 및 정수로 변환
					break
				}
			}
		}
	}

	results <- SubscriberResult{
		ID:           id,
		ReceivedSock: received,
		Expected:     expected,
		IsSuccessful: received == expected,
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
	for i := 1; i <= *sn; i++ {
		wg.Add(1)

		// 소켓 연결 (포트 p가 제공된 경우)
		var conn net.Conn

		if *port != "" {
			var err error
			// Dial을 사용해 서버에 연결
			conn, err = net.Dial("tcp", "localhost:"+*port) // The Dial function connects to a server:
			if err != nil {
				log.Fatalf("Failed to connect to publisher via socket: %v", err)
			}
			fmt.Printf("-- Subscriber %d connected to socket.\n", i)
		}
		// 구독자가 끝나면 소켓 연결 닫기
		defer conn.Close() // Close closes the connection.

		// 총 기대 메시지 수 = 메시지 발행 횟수 (n) × 구독자 수 (sn)
		// -> 각 구독자는 n번의 메시지를 받는 것이 목표이기 때문에, 각 구독자의 기대 메시지 수는 발행자의 n
		go subscribeToMQTT(client, *topic, i, &wg, results)

		if conn != nil {
			wg.Add(1)
			go receiveFromSocket(conn, i, &wg, results)
		}
	}

	go func() {
		wg.Wait()      // 모든 수신이 완료될 때까지 대기, Wait blocks until the WaitGroup counter is zero.
		close(results) // 모든 작업이 끝난 후 채널 닫기, close : shutting down the channel after the last sent value is received
	}()

	// 결과 출력
	for i := 0; i < *sn; i++ { // 구독자 수만큼 결과 수신
		result := <-results
		if result.IsSuccessful {
			fmt.Printf("=> Subscriber %d: Successfully received %d/%d messages.\n",
				result.ID, result.ReceivedMQTT, result.Expected)
		} else {
			fmt.Printf("=> Subscriber %d: Failed to receive all messages (%d/%d).\n",
				result.ID, result.ReceivedMQTT, result.Expected)
		}
	}
}
