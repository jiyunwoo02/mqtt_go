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

func publishMessages(client mqtt.Client, topic string, message string, n int, qos int, socketConn net.Conn) {
	for i := 1; i <= n; i++ {
		msg := fmt.Sprintf("%s#%d", message, i) // Hello#1 형식
		token := client.Publish(topic, byte(qos), false, msg)
		token.Wait()
		fmt.Printf("Published: %s\n", msg)

		// 소켓을 통해 메시지 전송
		_, err := socketConn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Printf("Error sending message via socket: %v", err)
		}

		time.Sleep(1 * time.Second) // 발행 간격
	}
	fmt.Println("All messages published.")
}

func main() {
	id := flag.String("id", "publisher1", "The id of the publisher")
	qos := flag.Int("qos", 0, "QoS level (0, 1, 2)")
	topic := flag.String("topic", "test/topic", "MQTT topic")
	n := flag.Int("n", 1, "Number of messages to publish") // n을 입력 받는 걸로 바꾸기
	address := flag.String("address", "tcp://localhost:1883", "Address of the broker")

	flag.Parse()

	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// // 사용자 입력 받기
	// fmt.Print("Enter the message to publish: ")
	// reader := bufio.NewReader(os.Stdin)
	// message, _ := reader.ReadString('\n')
	// message = strings.TrimSpace(message)

	// 소켓 서버 설정
	listener, err := net.Listen("tcp", "localhost:9090") // listen : crate server
	if err != nil {
		log.Fatalf("Failed to start socket server: %v", err)
	}
	defer listener.Close()

	fmt.Println("Waiting for subscriber connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Failed to accept subscriber connection: %v", err)
	}
	fmt.Println("Successed to accept subscriber")
	defer conn.Close()

	// 메시지 발행 시작
	// go publishMessages(client, *topic, message, *n, *qos, conn)

	select {} // 프로그램이 종료되지 않도록 대기
}
