// package main

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// // 메시지 수신 핸들러
// var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
// }

// func main() {
// 	// 첫 번째 브로커에 연결하는 발행자1 클라이언트 옵션 설정 (포트 1883)
// 	publisherOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
// 		SetClientID("publisherClient1").
// 		SetUsername("username").
// 		SetPassword("password")

// 	// 첫 번째 발행자 클라이언트 생성
// 	publisherClient1 := mqtt.NewClient(publisherOpts)
// 	if token := publisherClient1.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("발행자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자 %s 가 브로커 [tcp://localhost:1883]에 연결됨\n", publisherOpts.ClientID)

// 	// 두 번째 브로커에 연결하는 발행자2 클라이언트 옵션 설정 (포트 1884)
// 	publisherOpts2 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1884"). // 두 번째 브로커 주소
// 		SetClientID("publisherClient2").
// 		SetUsername("username").
// 		SetPassword("password")

// 	// 두 번째 발행자 클라이언트 생성
// 	publisherClient2 := mqtt.NewClient(publisherOpts2)
// 	if token := publisherClient2.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("발행자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자 %s 가 브로커 [tcp://localhost:1884]에 연결됨\n", publisherOpts2.ClientID)

// 	// 구독자 클라이언트 옵션 설정 (포트 1883)
// 	subscriberOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
// 		SetClientID("subscriberClient").
// 		SetDefaultPublishHandler(msgHandler)

// 	// 구독자 클라이언트 생성
// 	subscriberClient := mqtt.NewClient(subscriberOpts)
// 	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자 %s 가 브로커 [tcp://localhost:1883]에 연결됨\n", subscriberOpts.ClientID)

// 	// test/topic 구독
// 	if token := subscriberClient.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자 %s 가 test/topic 구독 완료\n", subscriberOpts.ClientID)
// 	}

// 	// 발행자 클라이언트 1이 메시지 발행
// 	message := "Hello, MQTT from Client 1!"
// 	token := publisherClient1.Publish("test/topic", 0, false, message)
// 	token.Wait()
// 	fmt.Printf("발행자 %s 가 메시지 발행: %s\n", publisherOpts.ClientID, message)

// 	// 발행자 클라이언트 2가 메시지 발행
// 	message2 := "Hello, MQTT from Client 2!"
// 	token2 := publisherClient2.Publish("test/topic", 0, false, message2)
// 	token2.Wait()
// 	fmt.Printf("발행자 %s 가 메시지 발행: %s\n", publisherOpts2.ClientID, message2)

// 	// 3초 동안 대기하여 메시지 수신 대기
// 	time.Sleep(3 * time.Second)

// 	// 클라이언트 종료
// 	subscriberClient.Disconnect(250)
// 	publisherClient1.Disconnect(250)
// 	publisherClient2.Disconnect(250)

// 	fmt.Println("클라이언트 종료됨")
// }

// /*
// 동일한 주제를 구독하고 있더라도 서로 다른 브로커에 연결된 클라이언트들은 서로의 메시지를 수신하지 못한다.
// 클라이언트가 메시지를 받기 위해서는 발행자와 구독자가 같은 브로커에 연결되어 있어야 한다.

// -> 이를 해결하기 위해서는 브로커 간의 브릿지(bridge) 설정을 통해 브로커 간 메시지를 공유하도록 설정할 수도 있다.
// 하지만 이는 기본 설정에서는 제공되지 않으며, 추가적인 설정이 필요하다.
// */
