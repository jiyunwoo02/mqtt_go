package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func publishMessages(client mqtt.Client, topic string, message string, n int, qos int, socketConn net.Conn) {
	for i := 1; i <= n; i++ {
		msg := fmt.Sprintf("%s#%d", message, i)
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
	qos := flag.Int("qos", 0, "QoS level (0, 1, 2)")
	topic := flag.String("topic", "test/topic", "MQTT topic")
	n := flag.Int("n", 1, "Number of messages to publish")
	address := flag.String("address", "tcp://localhost:1883", "Address of the broker")

	flag.Parse()

	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID("publisher")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 사용자 입력 받기
	fmt.Print("Enter the message to publish: ")
	reader := bufio.NewReader(os.Stdin)
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	// 소켓 서버 설정
	listener, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		log.Fatalf("Failed to start socket server: %v", err)
	}
	defer listener.Close()

	fmt.Println("Waiting for subscriber connection...")
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Failed to accept subscriber connection: %v", err)
	}
	defer conn.Close()

	// 메시지 발행 시작
	go publishMessages(client, *topic, message, *n, *qos, conn)

	select {} // 프로그램이 종료되지 않도록 대기
}
