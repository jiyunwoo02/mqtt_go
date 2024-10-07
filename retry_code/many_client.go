package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/* 여기에서 수신자가 누구인지 clientid를 출력을 못할까? -> ClientOptionsReader가 있다
ClientOptionsReader provides an interface for reading ClientOptions after the client has been initialized.
= 클라이언트가 초기화된 이후에 사용할 수 있다
주석 코드 실행하면 cannot call pointer method ClientID on ~ .ClientOptionsReader
*/

// 메시지 수신 핸들러
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// cliID := client.OptionsReader().ClientID()
	fmt.Printf("주제1에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

var msgHandler2 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("주제2에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

func main() {
	// 첫 번째 브로커에 연결하는 발행자1 클라이언트 옵션 설정 (포트 1883)
	publisherOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
		SetClientID("publisherClient1").
		SetUsername("username").
		SetPassword("password")

	// 첫 번째 발행자 클라이언트 생성
	publisherClient1 := mqtt.NewClient(publisherOpts)
	if token := publisherClient1.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자1 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())

	// 두 번째 브로커에 연결하는 발행자2 클라이언트 옵션 설정 (포트 1884)
	publisherOpts2 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884"). // 두 번째 브로커 주소
		SetClientID("publisherClient2").
		SetUsername("username").
		SetPassword("password")

	// 두 번째 발행자 클라이언트 생성
	publisherClient2 := mqtt.NewClient(publisherOpts2)
	if token := publisherClient2.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자2 %s 가 브로커 [%s]에 연결됨\n", publisherOpts2.ClientID, publisherOpts2.Servers[0].String())

	// 구독자1 클라이언트 옵션 설정 (포트 1883)
	subscriberOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
		SetClientID("subscriberClient")
		//SetDefaultPublishHandler(msgHandler)

	// 구독자1 클라이언트 생성
	subscriberClient := mqtt.NewClient(subscriberOpts)
	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자1 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

	// 구독자2 클라이언트 옵션 설정 (포트 1884)
	subscriberOpts2 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884").
		SetClientID("subscriberClient2")
		//SetDefaultPublishHandler(msgHandler2)

	// 구독자2 클라이언트 생성
	subscriberClient2 := mqtt.NewClient(subscriberOpts2)
	if token := subscriberClient2.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자1 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts2.ClientID, subscriberOpts2.Servers[0].String())

	// 구독자1 test/topic 구독
	if token := subscriberClient.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자1 %s 가 test/topic 구독 완료\n", subscriberOpts.ClientID)
	}

	// 구독자2 test/topic2 구독
	if token := subscriberClient2.Subscribe("test/topic2", 0, msgHandler2); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자2 %s 가 test/topic2 구독 완료\n", subscriberOpts2.ClientID)
	}

	// 발행자 클라이언트 1이 메시지 발행
	message := "Hello, MQTT from Client 1!"
	token := publisherClient1.Publish("test/topic", 0, false, message)
	token.Wait()
	fmt.Printf("발행자1 %s 가 메시지 발행: %s\n", publisherOpts.ClientID, message)

	// 발행자 클라이언트 2가 메시지 발행
	message2 := "Hello, MQTT from Client 2!"
	token2 := publisherClient2.Publish("test/topic2", 0, false, message2)
	token2.Wait()
	fmt.Printf("발행자2 %s 가 메시지 발행: %s\n", publisherOpts2.ClientID, message2)

	// 이미 주제에 대해 첫 메시지 발행한 이후!
	// 구독자3 클라이언트 옵션 설정 (포트 1884)
	subscriberOpts3 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884").
		SetClientID("subscriberClient3")

	// 구독자3 클라이언트 생성
	subscriberClient3 := mqtt.NewClient(subscriberOpts3)
	if token := subscriberClient3.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자3 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts3.ClientID, subscriberOpts3.Servers[0].String())

	// 구독자3 test/topic2 구독
	if token := subscriberClient3.Subscribe("test/topic2", 0, msgHandler2); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자3 %s 가 test/topic2 구독 완료\n", subscriberOpts3.ClientID)
	}

	// 발행자 클라이언트 2가 메시지 발행
	message3 := "Hello, MQTT from Client 2!"
	token3 := publisherClient2.Publish("test/topic2", 0, false, message3)
	token3.Wait()
	fmt.Printf("발행자2 %s 가 메시지 발행: %s\n", publisherOpts2.ClientID, message3)

	// 3초 동안 대기하여 메시지 수신 대기
	time.Sleep(3 * time.Second)

	// 클라이언트 종료
	subscriberClient.Disconnect(250)
	publisherClient1.Disconnect(250)
	publisherClient2.Disconnect(250)

	fmt.Println("클라이언트 종료됨")
}

/*
동일한 주제를 구독하고 있더라도 서로 다른 브로커에 연결된 클라이언트들은 서로의 메시지를 수신하지 못한다.
클라이언트가 메시지를 받기 위해서는 발행자와 구독자가 같은 브로커에 연결되어 있어야 한다.

-> 이를 해결하기 위해서는 브로커 간의 브릿지(bridge) 설정을 통해 브로커 간 메시지를 공유하도록 설정할 수도 있다.
하지만 이는 기본 설정에서는 제공되지 않으며, 추가적인 설정이 필요하다.
*/

/* 주제를 구독한 구독자들의 목록은 출력하지 못할까? -> 보안상의 문제로 못하는듯
In the MQTT protocol, there isn't a built-in mechanism for clients
	to directly obtain a list of other clients subscribed to a specific topic.

MQTT brokers generally do not provide this information to clients due to privacy and security concerns.
*/
