// // 얘는 결과 출력 안됨, 소켓 다 수신 안됨

// package main

// import (
// 	"bufio"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"net"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// // 각 구독자의 결과를 저장하는 구조체
// type Result struct {
// 	ID           int  // 구독자 클라이언트 아이디
// 	ReceivedMQTT int  // MQTT를 통해 받은 메시지 수
// 	ReceivedSock int  // 소켓을 통해 받은 메시지 수
// 	Expected     int  // 기대 메시지 수 (발행자의 n 값)
// 	IsSuccessful bool // 수신 성공 여부 (기대 메시지 수와 실제 수신 메시지 수 비교)
// }

// // MQTT와 소켓의 결과를 병합하고 저장하기 위한 맵 생성
// var subResults = make(map[int]*Result)

// // 동시성 문제 방지를 위한 뮤텍스
// var mu sync.Mutex

// // MQTT 메시지 수신 함수
// func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup) {
// 	defer wg.Done() // 구독 작업이 끝나면 작업 완료 처리
// 	defer mu.Unlock()

// 	// MQTT 주제 구독 및 메시지 수신 핸들러 설정 (QoS=2)
// 	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
// 		mu.Lock()

// 		payload := string(msg.Payload())
// 		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)

// 		result := subResults[id]
// 		result.ReceivedMQTT++

// 		// 첫 메시지에서 n 값을 추출하여 기대 메시지 수 설정
// 		if strings.Contains(payload, "n=") {
// 			// 발행된 메시지에서 "n=" 이 포함된 부분을 찾기 위해 문자열을 ", " 구분자로 분할
// 			parts := strings.Split(payload, ", ")

// 			// 분할된 부분을 순회하며, "n="으로 시작하는 부분을 탐색
// 			for _, part := range parts {
// 				if strings.HasPrefix(part, "n=") {
// 					// "n=" 접두어를 제거하여 숫자 값만 추출
// 					nValue := strings.TrimPrefix(part, "n=")

// 					// 추출된 값을 정수로 변환하여 expectedMessages에 저장
// 					subResults[id].Expected, _ = strconv.Atoi(nValue)

// 					// 추출된 n 값과 해당 구독자 ID를 출력
// 					// fmt.Printf("-> Subscriber %d: Extracted n value: %d\n", id, subResults[id].Expected)

// 					// n 값을 찾았으면 반복문 종료
// 					break
// 				}
// 			}
// 		}
// 	})
// 	time.Sleep(3 * time.Second) // 메시지 수신 대기
// }

// // 소켓 메시지 수신 함수
// // 발행자가 소켓을 통해 보낸 메시지를 구독자가 수신
// func receiveFromSocket(conn net.Conn, id int, wg *sync.WaitGroup) {
// 	defer wg.Done() // 이 함수가 끝나면 WaitGroup에서 작업 하나를 완료 처리
// 	defer mu.Unlock()

// 	// 소켓(데이터 통로) 연결된 conn에서 버퍼(소켓을 통해 전송된 데이터의 임시 저장 공간)로부터 데이터를 한 줄 단위로 읽는다
// 	scanner := bufio.NewScanner(conn) // 데이터를 읽는 즉시 버퍼에서 해당 데이터는 제거됨

// 	for scanner.Scan() {
// 		mu.Lock()

// 		message := scanner.Text()
// 		fmt.Printf("[Subscriber %d] Received from Socket: %s\n", id, message)

// 		result := subResults[id]
// 		result.ReceivedSock++

// 		// n 값을 추출해 기대 메시지 수 설정
// 		if strings.Contains(message, "n=") {
// 			parts := strings.Split(message, ", ")
// 			for _, part := range parts {
// 				// fmt.Printf("%s\n", part) -> , 을 기준으로 분할된다 : he 발행시 (n=1) -> he#1 과 n=3
// 				if strings.HasPrefix(part, "n=") {
// 					nValue := strings.TrimPrefix(part, "n=")
// 					subResults[id].Expected, _ = strconv.Atoi(nValue)
// 					break
// 				}
// 			}
// 		}
// 	}
// 	time.Sleep(3 * time.Second) // 메시지 수신 대기
// }

// func main() {
// 	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
// 	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
// 	id := flag.String("id", "subscriber1", "The id of the subscriber")
// 	topic := flag.String("tpc", "test/topic", "MQTT topic")
// 	sn := flag.Int("sn", 1, "Number of subscribers")                          // sn 플래그 추가 (sn : 구독자의 수)
// 	port := flag.String("p", "", "Port to connect for publisher connections") // p 플래그 추가 (p : 서버 포트, TCP 소켓 클라이언트 역할)

// 	flag.Parse() // 플래그 파싱

// 	// MQTT 클라이언트 설정
// 	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
// 	client := mqtt.NewClient(opts)
// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("Failed to connect to broker: %v", token.Error())
// 	}
// 	// defer client.Disconnect(250)

// 	// 구독자 생성 및 수신
// 	var wg sync.WaitGroup

// 	// 각 구독자의 결과를 초기화하고 저장
// 	for i := 1; i <= *sn; i++ {
// 		subResults[i] = &Result{ID: i}
// 		wg.Add(1)

// 		// 소켓 연결 (포트 p가 제공된 경우)
// 		var conn net.Conn

// 		if *port != "" {
// 			var err error
// 			// Dial을 사용해 서버에 연결
// 			conn, err = net.Dial("tcp", "localhost:"+*port) // The Dial function connects to a server:
// 			if err != nil {
// 				log.Fatalf("Failed to connect to publisher via socket: %v", err)
// 			}
// 			fmt.Printf("-- Subscriber %d connected to socket.\n", i)
// 		}

// 		defer conn.Close() // 구독자가 끝나면 소켓 연결 닫기

// 		// -> 각 구독자는 n번의 메시지를 받는 것이 목표이기 때문에, 각 구독자의 기대 메시지 수는 발행자의 n
// 		go subscribeToMQTT(client, *topic, i, &wg)

// 		if conn != nil {
// 			wg.Add(1)
// 			go receiveFromSocket(conn, i, &wg)
// 		}
// 	}

// 	wg.Wait()

// 	// 결과 출력
// 	for i := 1; i <= *sn; i++ { // 구독자 수만큼 결과 수신
// 		result := subResults[i]
// 		if result.ReceivedMQTT == result.Expected && result.ReceivedSock == result.Expected {
// 			fmt.Printf("=> Subscriber %d: Successfully received all messages (MQTT: %d/%d, Socket: %d/%d).\n",
// 				result.ID, result.ReceivedMQTT, result.Expected, result.ReceivedSock, result.Expected)
// 		} else {
// 			fmt.Printf("=> Subscriber %d: Failed to receive all messages (MQTT: %d/%d, Socket: %d/%d).\n",
// 				result.ID, result.ReceivedMQTT, result.Expected, result.ReceivedSock, result.Expected)
// 		}
// 	}
// }

// /*
// [코드 전체 로직]

// - MQTT 구독자(Subscriber) 역할을 수행
// - MQTT 브로커와 소켓 연결을 통해 발행자의 메시지를 구독하고, 받은 메시지의 개수를 확인 및 검증
// - 구독자는 MQTT와 소켓 양쪽에서 메시지를 수신하고, 이를 정리하여 결과를 출력

// ---

// 1. 명령행 플래그 설정 및 파싱
// - 명령행 인자로 MQTT 브로커 주소, 구독자 ID, 주제, 구독자 수(sn), 소켓 포트(p)를 입력
// - 소켓 포트가 없을 경우 소켓 연결을 건너뜀

// ---

// 2. MQTT 브로커 연결 설정
// - 입력받은 브로커 주소와 구독자 ID를 사용해 MQTT 브로커에 연결
// - 연결 실패 시 프로그램이 종료

// ---

// 3. 소켓 연결 (옵션)
// - 소켓 서버의 포트가 제공된 경우에만 소켓 연결을 시도
// - 발행자와의 소켓 연결이 성공하면, 소켓 메시지를 수신

// ---

// 4. 구독자 생성 및 메시지 수신
// - 구독자 수(sn) 만큼 구독자를 생성하고, MQTT와 소켓을 통해 메시지를 수신
// - `WaitGroup`과 고루틴을 사용해 여러 구독자가 동시에 실행

// ---

// 5. MQTT 메시지 수신 함수
// - MQTT 주제에 구독하며 메시지를 수신
// - 첫 번째 메시지에서 n 값을 추출해 기대 메시지 수를 설정
// - 수신한 메시지의 개수를 확인하고, 그 결과를 채널에 저장

// ---

// 6. 소켓 메시지 수신 함수
// - 소켓을 통해 메시지를 수신
// - 첫 메시지에서 n 값을 추출해 기대 메시지 수를 설정

// ---

// 7. 결과 출력 및 종료
// - 모든 구독자의 메시지 수신이 완료되면 결과를 출력
// - 각 구독자가 정상적으로 메시지를 수신했는지 여부를 확인

// ---

// 결과 요약
// 1. 명령행 인자로 설정된 MQTT 주제와 포트에 따라 구독자가 생성
// 2. 구독자는 MQTT와 소켓 양쪽에서 메시지를 수신
// 3. 각 구독자는 첫 번째 메시지에서 n 값을 추출해 기대 메시지 수를 설정
// 4. 모든 구독자가 메시지를 수신한 후, 성공 여부를 출력

// */
