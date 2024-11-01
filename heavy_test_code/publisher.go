package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net" // Go에서 소켓 통신을 활용하기 위해서는 net 패키지 사용
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 메시지 발행 함수: MQTT 브로커와 소켓을 통해 메시지 전송
func publishMessages(client mqtt.Client, topic string, message string, n int, qos int, socketConn net.Conn) {
	// 발행 횟수를 topic/count 주제로 먼저 발행 -> socket을 사용하지 않은 경우에도 수신자가 발행자의 메시지 발행 횟수를 알 수 있도록!
	countTopic := topic + "/count"
	token := client.Publish(countTopic, byte(qos), false, fmt.Sprintf("%d", n))
	token.Wait()
	fmt.Printf("-> Published count on %s: %d\n", countTopic, n)

	// 1. MQTT 브로커에 메시지 발행
	for i := 1; i <= n; i++ {
		// Sprintf: 형식화된 결과를 문자열로 반환 (예: Hello#1, Hello#2 ...) <-> Printf: 반환값이 없으며, 바로 콘솔에 출력
		msg := fmt.Sprintf("%s#%d", message, i) // Hello#1 형식 <- 메시지에 발행 순서 포함

		token := client.Publish(topic, byte(qos), false, msg) // retain=false
		token.Wait()                                          // 발행이 완료될 때까지 대기
		fmt.Printf("- Published: %s\n", msg)

		time.Sleep(1 * time.Second) // 발행 간격 조정 (1sec)
	}

	// 2. 소켓을 통해 메시지 전송 (옵션)
	if socketConn != nil {
		socketMessage := fmt.Sprintf("%s, n=%d\n", message, n) // Hello, n=3 형식
		_, err := socketConn.Write([]byte(socketMessage))      // writes data to the connection.
		if err != nil {
			log.Printf("-- Error sending message via socket: %v", err)
		}
		fmt.Println("-- Socket message sent")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("All messages published.")
}

func main() {
	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
	id := flag.String("id", "publisher1", "The id of the publisher")
	topic := flag.String("tpc", "test/topic", "MQTT topic")
	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
	qos := flag.Int("q", 0, "QoS level (0, 1, 2)")                            // qos 플래그 추가 (0, 1, 2)
	n := flag.Int("n", 1, "Number of messages to publish")                    // n 플래그 추가 (n : 발행하는 메시지의 반복 발행 횟수)
	port := flag.String("p", "", "Port to listen for subscriber connections") // p 플래그 추가 (p : 리슨 포트, TCP 소켓 서버 역할) -> 해당 서버에 연결된 구독자에게 직접 메시지 전달, 포트 미제공 시 연결 X

	flag.Parse() // 플래그 파싱

	// QoS가 0, 1, 2가 아닌 값이 제공된다면?
	if *qos < 0 || *qos > 2 {
		// Fatalf is equivalent to [Printf] followed by a call to os.Exit(1).
		// -- Printf와 os.Exit를 조합한 형태로, 형식화된 메시지를 출력한 후 프로그램을 종료
		log.Fatalf("-- Invalid QoS value: %d. Allowed values are 0, 1, or 2.", *qos)
	}

	// MQTT 브로커에 연결 설정
	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("-- Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 소켓 서버 설정: 포트(-p)가 제공된 경우에만!
	// 1. 포트 제공시 -> 소켓 서버 시작 및 구독자 연결 대기
	if *port != "" {
		// 1) 서버 소켓 생성: 지정된 포트에서 클라이언트 요청 대기

		// net.Listen(protocol, address) : tcp protocol, IP주소(localhost):포트(-p)
		// listener 객체 : 클라이언트의 연결을 대기, 서버 소켓 역할 수행, Accept() 호출해 클라이언트의 연결 요청 수락
		listener, err := net.Listen("tcp", "localhost:"+*port) // The Listen function creates servers:

		if err != nil {
			// %v : 해당 값에 맞는 기본 형식으로 출력해주는 역할 [에러 메시지를 있는 그대로 깔끔하게 출력]
			// err는 error 인터페이스 타입 -> %v 사용 -> 인터페이스가 담고 있는 에러 메시지가 기본 문자열 형식으로 출력됨
			log.Fatalf("-- Failed to start socket server on port %s: %v", *port, err)
		}
		defer listener.Close() // 함수 종료 시 리스너 자원 해제

		// 2) TCP 서버에서 listener 객체에 10초 타임아웃 설정
		timeoutDuration := 10 * time.Second
		// net.Listener 인터페이스 -> *net.TCPListener로 타입 변환 => SetDeadline: 특정 시간까지의 마감 시점을 설정하는 역할
		listener.(*net.TCPListener).SetDeadline(time.Now().Add(timeoutDuration))

		fmt.Printf("-- Waiting for subscriber connection on port %s...\n", *port)

		// 3) 연결 요청 대기
		conn, err := listener.Accept() // Accept waits for and returns the next connection to the listener.

		if err != nil {
			// 타임아웃 발생 시 에러 메시지와 함께 MQTT로 전환
			fmt.Println("-- Subscriber connection timeout. Switching to MQTT only.")
			conn = nil // MQTT 모드로 진행하기 위해 conn을 nil로 설정
		} else {
			fmt.Println("-- Subscriber connected on socket.")
		}

		// -- 소켓 서버에 구독자가 연결된 후에, 발행할 메시지를 입력 받도록 하자
		// 사용자로부터 발행할 메시지 입력 받기
		scanner := bufio.NewScanner(os.Stdin) // 표준 입력(키보드)을 줄 단위로 읽기 위한 스캐너 생성

		// 메시지 발행 후, 또 메시지를 입력받도록 함
		// exit 입력 시 발행 종료
		for {
			fmt.Print("Enter the message to publish (type 'exit' to quit): ")
			scanner.Scan()            // 사용자 입력을 대기하다가 엔터를 누르면 읽기 -> 읽고 내부 버퍼에 저장
			message := scanner.Text() // 읽은 입력을 문자열로 반환하여 변수에 저장.

			if strings.ToLower(message) == "exit" {
				// 사용자가 발행자 측에서 exit 입력 시 발행자 프로그램 종료
				// 수신자 측에서도 알 수 있게 함, 그 후 수신자 측에서 요약 결과 출력
				exitTopic := *topic + "/exit"
				client.Publish(exitTopic, 2, false, "exit")
				fmt.Println("-- Exiting publisher.")
				break
			}

			// 메시지 발행 시작 (MQTT와 소켓 모두로 전송)
			publishMessages(client, *topic, strings.TrimSpace(message), *n, *qos, conn)
		}
	}
}
