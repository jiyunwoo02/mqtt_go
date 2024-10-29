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
func publishMessages(client mqtt.Client, topic string, message string, n int, qos int, retain bool, socketConn net.Conn) {
	sendCount := 0 // 소켓으로 성공적으로 전송한 메시지 수를 기록
	for i := 1; i <= n; i++ {
		// Sprintf: 형식화된 결과를 문자열로 반환 (예: Hello#1, Hello#2 ...) <-> Printf: 반환값이 없으며, 바로 콘솔에 출력
		msg := fmt.Sprintf("%s#%d, n=%d", message, i, n) // Hello#1 형식 <- 메시지에 발행 순서와 총 발행 횟수 포함

		// // 1. MQTT 브로커에 메시지 발행
		token := client.Publish(topic, byte(qos), retain, msg)
		token.Wait() // 발행이 완료될 때까지 대기
		fmt.Printf("- Published: %s\n", msg)

		// 2. 소켓을 통해 메시지 전송 (옵션)
		if socketConn != nil {
			_, err := socketConn.Write([]byte(msg + "\n")) // writes data to the connection.
			if err != nil {
				log.Printf("Error sending message via socket: %v", err)
			} else {
				sendCount++ // 소켓으로 메시지가 잘 전달될 때마다 sendCount 1씩 증가
			}
			time.Sleep(500 * time.Millisecond) // 소켓 닫기 전에 대기 시간 추가
		}

		time.Sleep(1 * time.Second) // 발행 간격 조정 (1sec)
	}

	// 3. 발행 결과 확인 및 검증
	// 총 메시지를 몇 번 발송했는지 확인 -> 사용자가 입력한 플래그 n과 일치해야 함!
	if socketConn != nil {
		fmt.Printf("-> 발행자가 소켓을 통해 총 %d번 메시지를 발송했습니다.\n", sendCount)
		if n != sendCount { // 소켓 전송 실패
			fmt.Print("-> 사용자가 요청한 발행 횟수 n과 실제 소켓 전송 횟수 불일치!\n")
		}
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

	// QoS가 0, 1, 2가 아닌 값이 제공된다면?
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
	// defer client.Disconnect(250)

	// 소켓 서버 설정: 포트(-p)가 제공된 경우에만!
	var conn net.Conn

	// 1. 포트 미제공시 -> 소켓 서버 연결 X
	if *port == "" {
		fmt.Printf("No port provided. No socket server connection will be established.\n")
		return
	}

	// 2. 포트 제공시 -> 소켓 서버 시작 및 구독자 연결 대기
	if *port != "" {
		// 1) 서버 소켓 생성: 지정된 포트에서 클라이언트 요청 대기

		// net.Listen(protocol, address) : tcp protocol, IP주소(localhost):포트(-p)
		// listener 객체 : 클라이언트의 연결을 대기, 서버 소켓 역할 수행, Accept() 호출해 클라이언트의 연결 요청 수락
		listener, err := net.Listen("tcp", "localhost:"+*port) // The Listen function creates servers:

		if err != nil {
			// %v : 해당 값에 맞는 기본 형식으로 출력해주는 역할 [에러 메시지를 있는 그대로 깔끔하게 출력]
			// err는 error 인터페이스 타입 -> %v 사용 -> 인터페이스가 담고 있는 에러 메시지가 기본 문자열 형식으로 출력됨
			log.Fatalf("Failed to start socket server on port %s: %v", *port, err)
		}
		defer listener.Close() // 함수 종료 시 리스너 자원 해제

		// 2) 클라이언트 연결 요청 대기
		fmt.Printf("Waiting for subscriber connection on port %s...\n", *port)

		// 3) 새로운 연결이 들어오면 해당 요청을 수락하고, 연결 객체 conn을 반환 (연결 실패 시 에러 반환)
		conn, err = listener.Accept() // Accept waits for and returns the next connection to the listener.
		if err != nil {
			log.Fatalf("Failed to accept subscriber connection: %v", err)
		}
		defer conn.Close() // 연결 종료 시 자원 해제

		fmt.Println("-- Subscriber connected.")
	}

	// -- 소켓 서버에 구독자가 연결된 후에, 발행할 메시지를 입력 받도록 하자
	// 사용자로부터 발행할 메시지 입력 받기
	scanner := bufio.NewScanner(os.Stdin) // 표준 입력(키보드)을 줄 단위로 읽기 위한 스캐너 생성
	fmt.Print("Enter the message to publish: ")
	scanner.Scan()            // 사용자 입력을 대기하다가 엔터를 누르면 읽기 -> 읽고 내부 버퍼에 저장
	message := scanner.Text() // 읽은 입력을 문자열로 반환하여 변수에 저장.

	// 메시지 앞에 r이 있을 경우 retain=True
	retain := false
	if strings.HasPrefix(message, "r") {
		retain = true
		message = strings.TrimPrefix(message, "r")
	}

	// 메시지 발행 시작 (MQTT와 소켓 모두로 전송)
	go publishMessages(client, *topic, strings.TrimSpace(message), *n, *qos, retain, conn)

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

/*
log.Printf와 fmt.Printf의 차이점
: fmt.Printf는 단순한 메시지 출력을 위해 사용하고,
: log.Printf는 로그나 에러 메시지 관리에 더 적합하다.

특징			fmt.Printf			log.Printf
출력목적		일반 메시지 출력	  로그 메시지 출력
시간정보포함	 X				     O
용도			콘솔 출력			 에러 처리, 디버깅, 로그 관리
출력대상		표준 출력 (stdout)   로그 (stdout 또는 파일)

예) fmt.Printf("Hello, %s!\n", "World") -> Hello, World!
예) log.Printf("An error occurred: %v", err) -> 2024/10/28 17:45:33 An error occurred: <error message>
*/

/*
log.FatalF는?
: 형식화된 메시지를 출력한 후 프로그램을 종료하는 함수
: 내부적으로 log.Printf를 사용해 메시지를 출력한 뒤, os.Exit(1)을 호출하여 프로그램을 비정상 종료

-> log.Printf와 동일하게 형식화된 메시지를 출력, 시간 정보 포함, 복구 불가능한 오류 발생 시 메시지 출력 후 프로그램 강제 종료

*/

/*

// 사용자로부터 발행할 메시지 입력 받기
scanner := bufio.NewScanner(os.Stdin) // 표준 입력(키보드)으로부터 데이터를 읽기 위해 스캐너 생성
fmt.Print("Enter the message to publish: ") // 입력 요청 메시지 출력
scanner.Scan() // 사용자가 입력할 때까지 대기, 입력 후 엔터 키 입력 시 데이터 읽음
message := scanner.Text() // 사용자가 입력한 내용을 문자열로 저장

1. os.Stdin (표준 입력): 키보드 입력을 받을 수 있는 표준 입력 스트림.
- 프로그램이 실행되는 동안 키보드로부터 입력을 받을 수 있는 표준 입력 스트림
- os.Stdin을 통해 사용자가 입력한 데이터를 읽는다.

2. bufio.Scanner (버퍼 기반 스캐너): 입력을 줄 단위로 처리하는 스캐너.
- os.Stdin에서 입력을 줄 단위로 읽기 위해 bufio.Scanner를 사용
- 엔터 키가 눌릴 때까지 입력을 대기하고, 입력된 데이터를 내부 버퍼에 저장

3. scanner.Scan() (사용자 입력 읽기):
- 사용자가 엔터 키를 누르면 Scan()이 입력된 데이터를 읽어 들인다.
- 데이터가 정상적으로 읽히면 true를, 실패하면 false를 반환

4. scanner.Text() (입력된 텍스트 가져오기):
- 사용자가 입력한 데이터를 문자열 형태로 반환
- 이 값을 message 변수에 저장


*/
