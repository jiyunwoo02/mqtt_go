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
// type SubscriberResult struct {
// 	ID           int
// 	ReceivedMQTT int // MQTT를 통해 받은 메시지 수
// 	ReceivedSock int // 소켓을 통해 받은 메시지 수
// 	Expected     int // 기대 메시지 수
// 	IsSuccessful bool
// }

// // MQTT 메시지 수신 함수
// func subscribeToMQTT(client mqtt.Client, topic string, id int, wg *sync.WaitGroup, results chan<- SubscriberResult) {
// 	// defer wg.Done()

// 	received := 0 // 수신한 메시지 개수
// 	expected := 0 // 발행자가 보낸 총 메시지 횟수(n). 첫 메시지에서 n 값을 추출하여 설정.

// 	// 메시지 수신 시 호출되는 콜백 함수: 메시지 핸들러
// 	// 메시지를 정확히 한 번 수신하도록 qos=2로 설정
// 	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
// 		payload := string(msg.Payload())
// 		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)

// 		// 첫 번째 메시지에서 n 값을 추출해 기대 메시지 수 설정
// 		if expected == 0 && strings.Contains(payload, "n=") {
// 			// 발행된 메시지에서 "n=" 이 포함된 부분을 찾기 위해 문자열을 ", " 구분자로 분할
// 			parts := strings.Split(payload, ", ")

// 			// 분할된 부분을 순회하며, "n="으로 시작하는 부분을 탐색
// 			for _, part := range parts {
// 				if strings.HasPrefix(part, "n=") {
// 					// "n=" 접두어를 제거하여 숫자 값만 추출
// 					nValue := strings.TrimPrefix(part, "n=")

// 					// 추출된 값을 정수로 변환하여 expectedMessages에 저장
// 					expected, _ = strconv.Atoi(nValue)

// 					// 추출된 n 값과 해당 구독자 ID를 출력
// 					fmt.Printf("=> Subscriber %d: Extracted n value: %d\n", id, expected)

// 					// n 값을 찾았으면 반복문 종료
// 					break
// 				}
// 			}
// 		} else {
// 			received++ // 메시지를 수신할 때마다 카운트 증가
// 		}
// 	})

// 	time.Sleep(100 * time.Second) // 메시지 수신 대기

// 	// 결과 저장
// 	results <- SubscriberResult{
// 		ID:           id,
// 		ReceivedMQTT: received,
// 		Expected:     expected,
// 		IsSuccessful: received == expected,
// 	}

// 	wg.Done()
// 	fmt.Println(time.Now())
// 	fmt.Print("Done 호출\n")
// }

// // 소켓 메시지 수신 함수
// // 발행자가 소켓을 통해 보낸 메시지를 구독자가 수신
// func receiveFromSocket(conn net.Conn, id int, wg *sync.WaitGroup, results chan<- SubscriberResult) {
// 	defer wg.Done() // 이 함수가 끝나면 WaitGroup에서 작업 하나를 완료 처리

// 	scanner := bufio.NewScanner(conn) // 소켓(데이터 통로) 연결된 conn에서 버퍼(소켓을 통해 전송된 데이터의 임시 저장 공간)로부터 데이터를 한 줄 단위로 읽는다, 데이터를 읽는 즉시 버퍼에서 해당 데이터는 제거됨
// 	received := 0                     // 수신한 메시지 개수를 저장하는 변수
// 	expected := 0                     // n 값 추출 후 저장될 변수

// 	for scanner.Scan() {
// 		message := scanner.Text()
// 		fmt.Printf("[Subscriber %d] Received from Socket: %s\n", id, message)

// 		// 첫 번째 메시지에서 n 값을 추출해 기대 메시지 수 설정
// 		if expected == 0 && strings.Contains(message, "n=") {
// 			parts := strings.Split(message, ", ")
// 			for _, part := range parts {
// 				// fmt.Printf("%s\n", part) -> , 을 기준으로 분할된다 : 사용자가 he 발행시 (n=1) -> he#1 과 n=3
// 				if strings.HasPrefix(part, "n=") {
// 					nValue := strings.TrimPrefix(part, "n=")
// 					expected, _ = strconv.Atoi(nValue) // n 값을 추출 및 정수로 변환
// 					fmt.Printf("=> Subscriber %d: Extracted n value: %d\n", id, expected)
// 					break
// 				}
// 			}
// 		} else {
// 			received++ // 메시지를 수신할 때마다 카운트 증가
// 		}
// 	}

// 	results <- SubscriberResult{
// 		ID:           id,
// 		ReceivedSock: received,
// 		Expected:     expected,
// 		IsSuccessful: received == expected,
// 	}
// }

// func main() {
// 	t := time.Now()
// 	fmt.Println(t)
// 	fmt.Print("\n")

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
// 	defer client.Disconnect(250)

// 	// 소켓 연결 (포트 p가 제공된 경우)
// 	var conn net.Conn
// 	if *port == "" {
// 		fmt.Printf("No port provided. No socket server connection will be established.")
// 	}
// 	if *port != "" {
// 		var err error
// 		conn, err = net.Dial("tcp", "localhost:"+*port) // The Dial function connects to a server:
// 		if err != nil {
// 			log.Fatalf("Failed to connect to publisher via socket: %v", err)
// 		}
// 		defer conn.Close() // Close closes the connection.
// 		fmt.Println("Connected to publisher via socket.")
// 	}

// 	// 구독자 생성 및 수신
// 	var wg sync.WaitGroup
// 	results := make(chan SubscriberResult, *sn) // 구독자 수만큼 결과 저장할 채널 생성

// 	for i := 1; i <= *sn; i++ {
// 		wg.Add(1)
// 		fmt.Print("Add 호출\n")
// 		// 총 기대 메시지 수 = 메시지 발행 횟수 (n) × 구독자 수 (sn)
// 		// -> 각 구독자는 n번의 메시지를 받는 것이 목표이기 때문에, 각 구독자의 기대 메시지 수는 발행자의 n
// 		go subscribeToMQTT(client, *topic, i, &wg, results)
// 		// fmt.Println(time.Now())

// 		if conn != nil {
// 			wg.Add(1)
// 			go receiveFromSocket(conn, i, &wg, results)
// 			// fmt.Println(time.Now())
// 		}
// 	}

// 	// 일부 고루틴이 아직 작업을 마치지 않았는데도 wg.Wait()가 완료되어 다음 단계로 넘어가 results 채널이 조기 종료되는 문제가 발생
// 	// subscribeToMQTT 고루틴이 예상보다 빨리 끝나서 wg.Done()을 호출해버리면, WaitGroup의 카운터는 1 감소
// 	// -- subscribeToMQTT와 receiveFromSocket 고루틴 중 일부가 wg.Done()을 호출해서 카운터를 줄이면, wg.Wait()는 남은 작업이 아직 실행 중임에도 WaitGroup 카운터가 0이 되면 종료
// 	// wg.Wait()가 끝나자마자 close(results)가 실행되어 results 채널을 조기에 닫게 되는데, 이로 인해 results <- SubscriberResult{...}와 같이 아직 results 채널에 기록되지 않은 고루틴 결과가 무시
// 	// -- 이로 인해 일부 구독자가 메시지를 수신하지 못한 채 대기 상태

// 	// wg.Wait()      // 모든 수신이 완료될 때까지 대기, Wait blocks until the WaitGroup counter is zero.
// 	// close(results) // close : shutting down the channel after the last sent value is received

// 	wg.Wait() // 모든 수신이 완료될 때까지 대기, Wait blocks until the WaitGroup counter is zero.
// 	fmt.Print("Wait 종료\n")
// 	close(results) // close : shutting down the channel after the last sent value is received
// 	fmt.Print("channel close\n")

// 	fmt.Print("for문 시작\n")
// 	// 결과를 읽어 처리
// 	for result := range results {
// 		fmt.Print("for문 순회\n")
// 		if result.IsSuccessful {
// 			fmt.Printf("Subscriber %d: Successfully received %d/%d messages.\n", result.ID, result.ReceivedMQTT, result.Expected)
// 		} else {
// 			fmt.Printf("Subscriber %d: Failed to receive all messages (%d/%d).\n", result.ID, result.ReceivedMQTT, result.Expected)
// 		}
// 	}

// 	fmt.Print("for문 종료\n")
// 	fmt.Println(time.Now())
// 	fmt.Printf("실행 시간: %s\n", time.Since(t).Seconds())

// 	// go func() {
// 	// 	wg.Wait()
// 	// 	// close(results)
// 	// }()

// 	// 1. 고루틴 안에 wg.Wait()와 close(results)
// 	// 소켓에 연결되었다는 메시지만 출력되고 프로그램이 바로 종료되는 문제
// 	// -> results 채널을 읽는 루프가 없기 때문
// 	// => results 채널에 들어오는 구독 결과를 읽어주지 않으면, 프로그램은 close(results)가 실행된 후 메인 고루틴이 종료되면서 바로 프로그램이 끝나버림

// 	// 2. 고루틴 안에 wg.Wait()
// 	// 소켓에 연결되었다는 메시지만 출력되고 프로그램이 바로 종료되는 문제
// 	// -> results 채널을 닫는 close(results)나 results를 처리하는 루프가 없어 프로그램이 종료
// 	// -> 현재 코드는 WaitGroup을 기다리기만 할 뿐, 그 이후에 데이터를 처리하지 않아 프로그램이 예상보다 일찍 종료됨
// 	// => Go 프로그램의 메인 고루틴이 종료 조건을 만족할 때 즉시 종료되기 때문, main() 함수가 완료되면 모든 고루틴이 즉시 중지
// 	// => results 채널을 생성했지만 그 채널에서 데이터를 읽어오는 루프가 없다면, results에 전달된 값이 처리되지 않은 상태로 남음
// 	// => results를 처리하는 루프가 없다면 main() 고루틴은 더 이상 기다릴 필요가 없기 때문에 종료됨

// 	// 해당 코드를 모든 수신자가 메시지를 수신하도록 만드려면?
// 	// 1. 결과 처리 고루틴 추가
// 	// : wg.Wait()가 완료된 후 close(results)가 호출되도록 고루틴으로 처리하여 모든 고루틴이 작업을 마칠 때까지 대기
// 	// 2. 채널의 데이터 읽기 루프 추가
// 	// : for result := range results를 통해 results 채널이 닫힐 때까지 데이터를 읽어 결과를 출력
// }

// /* 문제: sn=3이고, 소켓 연결 사용한다 가정

// 1. 메인 고루틴 시작
// - 총 3명의 구독자 고루틴 생성
// - 각 구독자는 MQTT와 소켓을 통해 메시지 수신

// 2. 반복문 통한 구독자 고루틴 생성
// - 반복문 총 3번 실행
// - 첫번째 반복:
// WaitGroup의 카운트가 1 증가 -> go subscribeToMQTT(...) 고루틴이 생성되어 MQTT를 통해 메시지를 수신 -> 소켓 연결이 있으므로 wg.Add(1)이 다시 호출되어 카운트가 1 증가
// -> go receiveFromSocket(...) 고루틴이 생성되어 소켓을 통해 메시지를 수신 -> 현재 WaitGroup 카운트는 2
// - 두번째 반복:
// wg.Add(1) 호출로 WaitGroup 카운트가 1 증가하여 현재 카운트는 3 -> go subscribeToMQTT(...) 고루틴이 생성 -> wg.Add(1)이 다시 호출되어 카운트가 4
// -> go receiveFromSocket(...) 고루틴이 생성 -> 현재 WaitGroup 카운트는 4
// - 세번째 반복:
// wg.Add(1)로 WaitGroup 카운트가 1 증가하여 카운트는 5 -> go subscribeToMQTT(...) 고루틴이 생성 -> wg.Add(1)이 다시 호출되어 카운트가 6
// -> go receiveFromSocket(...) 고루틴이 생성 -> 최종적으로 WaitGroup 카운트는 6

// 3. 모든 구독자 고루틴 실행 완료 후 대기
// - 메인 고루틴은 wg.Wait()에 도달하여 모든 고루틴이 완료될 때까지 대기
// - subscribeToMQTT 및 receiveFromSocket 고루틴이 완료될 때마다 wg.Done()이 호출되어 WaitGroup 카운트가 감소
// - WaitGroup의 카운트가 0이 될 때까지 메인 고루틴은 wg.Wait()에서 대기

// 4. 모든 고루틴 완료 후 채널 닫기
// - 모든 구독자 고루틴이 완료되어 WaitGroup 카운트가 0이 되면, wg.Wait()가 완료
// - close(results)가 호출되어 결과 채널이 닫힘


// */

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

// 결과 요약
// 1. 명령행 인자로 설정된 MQTT 주제와 포트에 따라 구독자가 생성
// 2. 구독자는 MQTT와 소켓 양쪽에서 메시지를 수신
// 3. 각 구독자는 첫 번째 메시지에서 n 값을 추출해 기대 메시지 수를 설정
// 4. 모든 구독자가 메시지를 수신한 후, 성공 여부를 출력

// */
