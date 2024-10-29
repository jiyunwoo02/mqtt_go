// // 얘는 결과 출력은 되는데 수신 안됨

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
// 	defer wg.Done() // 구독 작업이 끝나면 WaitGroup에서 작업 완료 처리

// 	client.Subscribe(topic, 2, func(_ mqtt.Client, msg mqtt.Message) {
// 		payload := string(msg.Payload())
// 		fmt.Printf("[Subscriber %d] Received from MQTT: %s\n", id, payload)

// 		mu.Lock()
// 		defer mu.Unlock()

// 		result := subResults[id]
// 		result.ReceivedMQTT++

// 		if result.Expected == 0 && strings.Contains(payload, "n=") {
// 			parts := strings.Split(payload, ", ")
// 			for _, part := range parts {
// 				if strings.HasPrefix(part, "n=") {
// 					nValue := strings.TrimPrefix(part, "n=")
// 					result.Expected, _ = strconv.Atoi(nValue)
// 					break
// 				}
// 			}
// 		}
// 	})
// 	time.Sleep(3 * time.Second) // 메시지 수신 대기
// }

// // 소켓 메시지 수신 함수
// func receiveFromSocket(conn net.Conn, id int, wg *sync.WaitGroup) {
// 	defer wg.Done() // 이 함수가 끝나면 WaitGroup에서 작업 완료 처리

// 	scanner := bufio.NewScanner(conn)
// 	for scanner.Scan() {
// 		message := scanner.Text()
// 		fmt.Printf("[Subscriber %d] Received from Socket: %s\n", id, message)

// 		mu.Lock()
// 		defer mu.Unlock()

// 		result := subResults[id]
// 		result.ReceivedSock++

// 		if result.Expected == 0 && strings.Contains(message, "n=") {
// 			parts := strings.Split(message, ", ")
// 			for _, part := range parts {
// 				if strings.HasPrefix(part, "n=") {
// 					nValue := strings.TrimPrefix(part, "n=")
// 					result.Expected, _ = strconv.Atoi(nValue)
// 					break
// 				}
// 			}
// 		}
// 	}
// 	time.Sleep(3 * time.Second)
// }

// func main() {
// 	address := flag.String("add", "tcp://localhost:1883", "Address of the broker")
// 	id := flag.String("id", "subscriber1", "The id of the subscriber")
// 	topic := flag.String("tpc", "test/topic", "MQTT topic")
// 	sn := flag.Int("sn", 1, "Number of subscribers")
// 	port := flag.String("p", "", "Port to connect for publisher connections")

// 	flag.Parse()

// 	opts := mqtt.NewClientOptions().AddBroker(*address).SetClientID(*id)
// 	client := mqtt.NewClient(opts)
// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("Failed to connect to broker: %v", token.Error())
// 	}
// 	defer client.Disconnect(250)

// 	var wg sync.WaitGroup

// 	for i := 1; i <= *sn; i++ {
// 		subResults[i] = &Result{ID: i}

// 		wg.Add(1)

// 		var conn net.Conn
// 		if *port != "" {
// 			var err error
// 			conn, err = net.Dial("tcp", "localhost:"+*port)
// 			if err != nil {
// 				log.Fatalf("Failed to connect to publisher via socket: %v", err)
// 			}
// 			fmt.Printf("-- Subscriber %d connected to socket.\n", i)

// 			go func(c net.Conn, id int) {
// 				defer c.Close()
// 				receiveFromSocket(c, id, &wg)
// 			}(conn, i)
// 		}

// 		go subscribeToMQTT(client, *topic, i, &wg)
// 	}

// 	wg.Wait()

// 	for i := 1; i <= *sn; i++ {
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
