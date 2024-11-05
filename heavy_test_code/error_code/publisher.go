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
	sendCount := 0 // 소켓으로 전송한 메시지 수를 카운트
	for i := 1; i <= n; i++ {
		msg := fmt.Sprintf("%s#%d, n=%d", message, i, n) // Hello#1 형식, Sprintf: 형식화된 결과를 문자열로 반환 <-> Printf: 반환값이 없으며, 바로 콘솔에 출력

		// MQTT 브로커에 메시지 발행
		token := client.Publish(topic, byte(qos), false, msg)
		token.Wait()
		fmt.Printf("- Published: %s\n", msg)

		// 소켓을 통해 메시지 전송
		if socketConn != nil {
			_, err := socketConn.Write([]byte(msg + "\n")) // writes data to the connection.
			if err != nil {
				log.Printf("Error sending message via socket: %v", err)
			} else {
				sendCount++ // 소켓으로 메시지가 잘 전달될 때마다 sendCount 1씩 증가
			}
		}

		time.Sleep(1 * time.Second) // 발행 간격
	}

	// 총 메시지를 몇 번 발송했는지 확인 -> 사용자가 입력한 플래그 n과 일치해야 함!
	if socketConn != nil {
		fmt.Printf("-> 발행자가 소켓을 통해 총 %d번 메시지를 발송했습니다.\n", sendCount)
	}

	if n != sendCount {
		fmt.Print("-> 사용자가 원한 발행 횟수 n과 실제 발행 횟수 불일치!")
	}

	fmt.Println("All messages published.")
}

// 발행자
func main() {
	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
	id := flag.String("id", "publisher1", "The id of the publisher")
	topic := flag.String("tpc", "test/topic", "MQTT topic")
	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
	qos := flag.Int("q", 0, "QoS level (0, 1, 2)")                            // qos 플래그 추가 (0, 1, 2)
	n := flag.Int("n", 1, "Number of messages to publish")                    // n 플래그 추가 (n : 발행하는 메시지의 반복 발행 횟수)
	port := flag.String("p", "", "Port to listen for subscriber connections") // p 플래그 추가 (p : 리슨 포트, TCP 소켓 서버 역할) -> 해당 서버에 연결된 구독자에게 직접 메시지 전달, 포트 미제공 시 연결 X

	flag.Parse() // 플래그 파싱

	// QoS가 0, 1, 2가 아닌 값이 제공된다면? -> QoS
	if *qos < 0 || *qos > 2 {
		// Fatalf is equivalent to [Printf] followed by a call to os.Exit(1).
		// -- Printf와 os.Exit를 조합한 형태로, 형식화된 메시지를 출력한 후 프로그램을 종료
		log.Fatalf("Invalid QoS value: %d. Allowed values are 0, 1, or 2.", *qos)
	}

	// MQTT 브로커에 연결 설정
	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 소켓 서버 설정: 포트(-p)가 제공된 경우에만!
	var conn net.Conn
	if *port == "" {
		fmt.Printf("No port provided. No socket server connection will be established.\n")
	}
	if *port != "" {
		listener, err := net.Listen("tcp", "localhost:"+*port) // The Listen function creates servers:
		if err != nil {
			log.Fatalf("Failed to start socket server: %v", err)
		}
		defer listener.Close()

		fmt.Printf("Waiting for subscriber connection on port %s...\n", *port)
		conn, err = listener.Accept() // Accept waits for and returns the next connection to the listener.
		if err != nil {
			log.Fatalf("Failed to accept subscriber connection: %v", err)
		}
		defer conn.Close()

		fmt.Println("-- Subscriber connected.")
	}

	// 소켓 서버에 구독자가 연결된 후에, 발행할 메시지를 입력 받도록 하자
	// 사용자로부터 발행할 메시지 입력 받기
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the message to publish: ")
	scanner.Scan()
	message := scanner.Text()

	// 메시지 발행 시작 (MQTT와 소켓 모두로 전송)
	go publishMessages(client, *topic, strings.TrimSpace(message), *n, *qos, conn)

	select {} // 프로그램이 종료되지 않도록 대기
}

/*
[코드 전체 로직]

- MQTT 발행자(Publisher) 역할을 수행하며, MQTT 브로커와 소켓 연결을 통해 메시지를 발행하고 전달
- 사용자는 명령행 플래그로 발행 설정을 입력하고 메시지를 반복 발행

---

1. 명령행 플래그 파싱
- 발행자 ID, QoS 값(0~2), 발행할 메시지 횟수, 소켓 리슨 포트를 명령행에서 입력받아 설정
- 잘못된 QoS 값이 입력되면 프로그램 종료

---

2. MQTT 브로커 연결
- MQTT 브로커와 연결해 발행할 준비

---

3. 소켓 서버 설정 (옵션)
- 소켓 서버는 제공된 포트로 구독자와 연결
- 포트가 없으면 소켓 서버는 시작하지 않음

---

4. 사용자 입력 메시지 수집 및 처리
- 사용자가 발행할 메시지를 입력받고, `r`로 시작하면 retain 플래그를 활성화

---

5. 메시지 발행
- MQTT 브로커와 소켓을 통해 메시지를 반복 발행
- 각 메시지는 `<메시지>#<순서>, n=<반복횟수>` 형식으로 발행

---

6. 프로그램 실행 유지
- 프로그램이 종료되지 않고 구독자와의 연결을 유지

---

결과 요약
- MQTT 브로커와 소켓 연결을 통한 메시지 발행이 수행
- 사용자는 메시지 내용과 반복 횟수를 입력하며, 메시지가 발행될 때마다 콘솔에 출력
- 구독자가 연결된 경우, 소켓으로도 동일한 메시지를 전달
*/
