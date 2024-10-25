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
	sendCount := 0 // 소켓으로 전송한 메시지 수를 카운트
	for i := 1; i <= n; i++ {
		msg := fmt.Sprintf("%s#%d", message, i) // Hello#1 형식

		// MQTT 브로커에 메시지 발행
		token := client.Publish(topic, byte(qos), retain, msg)
		token.Wait()
		fmt.Printf("Published: %s\n", msg)

		// 소켓을 통해 메시지 전송
		if socketConn != nil {
			_, err := socketConn.Write([]byte(msg + "\n")) // writes data to the connection.
			if err != nil {
				log.Printf("Error sending message via socket: %v", err)
			} else {
				sendCount++ // 소켓으로 메시지가 잘 전달될 때마다 i 1씩 증가
			}
		}

		time.Sleep(1 * time.Second) // 발행 간격
	}
	if socketConn != nil {
		fmt.Printf("발행자가 소켓을 통해 총 %d번 메시지를 발송했습니다.\n", sendCount)
	} // 총 메시지를 몇 번 발송했는지 확인 -> n과 일치해야 함!

	fmt.Println("All messages published.")
}

func main() {
	// 명령행 플래그 설정 (플래그명, 기본값, 설명)
	id := flag.String("id", "publisher1", "The id of the publisher")
	qos := flag.Int("q", 0, "QoS level (0, 1, 2)")
	topic := flag.String("tp", "test/topic", "MQTT topic")
	n := flag.Int("n", 1, "Number of messages to publish")
	address := flag.String("ad", "tcp://localhost:1883", "Address of the broker")
	port := flag.String("p", "", "Port to listen for subscriber connections") // 포트 미제공 시 연결 X

	flag.Parse() // 플래그 파싱

	// MQTT 브로커에 연결 설정
	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 사용자로부터 메시지 입력 받기
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the message to publish: ")
	scanner.Scan()
	message := scanner.Text()

	// 소켓 서버 설정: 포트(-p)가 제공된 경우에만!
	var conn net.Conn
	if *port != "" {
		listener, err := net.Listen("tcp", "localhost:"+*port)
		if err != nil {
			log.Fatalf("Failed to start socket server: %v", err)
		}
		defer listener.Close()

		fmt.Printf("Waiting for subscriber connection on port %s...\n", *port)
		conn, err = listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept subscriber connection: %v", err)
		}
		defer conn.Close()
		fmt.Println("Subscriber connected.")
	}

	// 메시지 발행 루프: 앞에 r이 있을 경우 retain=True
	retain := false
	if strings.HasPrefix(message, "r") {
		retain = true
		message = strings.TrimPrefix(message, "r")
	}

	// 메시지 발행 시작 (MQTT와 소켓 모두로 전송)
	go publishMessages(client, *topic, strings.TrimSpace(message), *n, *qos, retain, conn)

	select {} // 프로그램이 종료되지 않도록 대기
}
