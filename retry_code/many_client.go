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

// MQTT는 비동기적으로 작동하기 때문에, 메시지 순서가 변경되어 출력되는 경우 발생!
// -> 메시지 처리를 동기적으로 만들거나, 메시지 발행과 수신 사이에 약간의 지연 시간을 두는 방법을 사용
// => 네트워크 지연이나 브로커 처리 속도에 따라 메시지 순서가 달라질 수 있음을 의미
// => 따라서 비동기 통신에서는 항상 순서를 보장할 수 없으며, 필요한 경우 동기화 메커니즘을 추가해야 함

// clean session 구독자의 모든 세션 정보 저장 여부 (기본 false)
// true -> 구독자가 연결을 끊으면 그 구독자의 모든 세션 정보가 브로커에서 삭제

// 메시지 수신 핸들러
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// cliID := client.OptionsReader().ClientID()
	fmt.Printf("주제1에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

var msgHandler2 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("주제2에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

var msgHandler3 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("주제3에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
}

func main() {
	// 첫 번째 브로커에 연결하는 발행자1 클라이언트 옵션 설정 (포트 1883)
	publisherOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
		SetClientID("publisher1").
		SetCleanSession(true) // 클린 세션 활성화
		// SetUsername("username1").
		// SetPassword("password1")

	// 첫 번째 발행자 클라이언트 생성
	publisherClient1 := mqtt.NewClient(publisherOpts)
	if token := publisherClient1.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자1 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())

	// 브로커에 이미 retain된 메시지를 삭제 -> 빈 메시지를 retain 플래그와 함께 발행
	token := publisherClient1.Publish("test/topic", 0, true, "")
	token.Wait()

	// 두 번째 브로커에 연결하는 발행자2 클라이언트 옵션 설정 (포트 1884)
	publisherOpts2 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884"). // 두 번째 브로커 주소
		SetClientID("publisher2").
		SetCleanSession(true)
		// SetUsername("username2").
		// SetPassword("password2")

	// 두 번째 발행자 클라이언트 생성
	publisherClient2 := mqtt.NewClient(publisherOpts2)
	if token := publisherClient2.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자2 %s 가 브로커 [%s]에 연결됨\n", publisherOpts2.ClientID, publisherOpts2.Servers[0].String())

	// 구독자1 클라이언트 옵션 설정 (포트 1883)
	subscriberOpts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
		SetClientID("subscriber1").
		SetDefaultPublishHandler(msgHandler).
		SetCleanSession(true)

	// 구독자1 클라이언트 생성
	subscriberClient := mqtt.NewClient(subscriberOpts)
	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자1 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

	// 구독자2 클라이언트 옵션 설정 (포트 1884)
	subscriberOpts2 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884").
		SetClientID("subscriber2").
		SetDefaultPublishHandler(msgHandler2).
		SetCleanSession(true)

	// 구독자2 클라이언트 생성
	subscriberClient2 := mqtt.NewClient(subscriberOpts2)
	if token := subscriberClient2.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자2 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts2.ClientID, subscriberOpts2.Servers[0].String())

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
	message := "pub1_msg1"
	token1 := publisherClient1.Publish("test/topic", 0, false, message)
	token1.Wait()
	fmt.Printf("발행자1 %s 가 메시지1 발행: %s\n", publisherOpts.ClientID, message)
	time.Sleep(3 * time.Second) // 메시지 전송 후 대기

	// 발행자 클라이언트 2가 메시지 발행
	message2 := "pub2_msg1"
	token2 := publisherClient2.Publish("test/topic2", 0, false, message2)
	token2.Wait()
	fmt.Printf("발행자2 %s 가 메시지1 발행: %s\n", publisherOpts2.ClientID, message2)
	time.Sleep(3 * time.Second) // 메시지 전송 후 대기

	// 이미 주제에 대해 첫 메시지 발행한 이후!
	// 구독자3 클라이언트 옵션 설정 (포트 1884)
	subscriberOpts3 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1884").
		SetClientID("subscriber3").
		SetDefaultPublishHandler(msgHandler2).
		SetCleanSession(true)

	// 구독자3 클라이언트 생성
	subscriberClient3 := mqtt.NewClient(subscriberOpts3)
	if token := subscriberClient3.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 3 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자3 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts3.ClientID, subscriberOpts3.Servers[0].String())

	// 구독자3 test/topic2 구독
	if token := subscriberClient3.Subscribe("test/topic2", 0, msgHandler2); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자3 %s 가 test/topic2 구독 완료\n", subscriberOpts3.ClientID)
	}

	// 발행자 클라이언트 2가 메시지 발행
	message3 := "pub2_msg2"
	token3 := publisherClient2.Publish("test/topic2", 0, false, message3)
	token3.Wait()
	fmt.Printf("발행자2 %s 가 메시지2 발행: %s\n", publisherOpts2.ClientID, message3)
	time.Sleep(3 * time.Second) // 메시지 전송 후 대기

	// // 발행자 클라이언트 1가 메시지 발행
	// message4 := "pub1_msg2_retain"
	// token4 := publisherClient1.Publish("test/topic", 0, true, message4)
	// token4.Wait()
	// fmt.Printf("발행자1 %s 가 메시지2 발행 (retain): %s\n", publisherOpts.ClientID, message4)
	// time.Sleep(3 * time.Second) // 메시지 전송 후 대기

	// // 구독자4가 주제가 이미 메시지를 발행한 후에 구독 -> retain=true를 통해 전달받음을 확인
	// // 구독자4 클라이언트 옵션 설정 (포트 1883)
	// subscriberOpts4 := mqtt.NewClientOptions().
	// 	AddBroker("tcp://localhost:1883").
	// 	SetClientID("subscriber4").
	// 	SetCleanSession(true)

	// // 구독자4 클라이언트 생성
	// subscriberClient4 := mqtt.NewClient(subscriberOpts4)
	// if token := subscriberClient4.Connect(); token.Wait() && token.Error() != nil {
	// 	log.Fatalf("구독자 클라이언트 4 브로커에 연결 실패: %v\n", token.Error())
	// }
	// fmt.Printf("구독자4 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts4.ClientID, subscriberOpts4.Servers[0].String())

	// // 구독자4 test/topic 구독
	// if token := subscriberClient4.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil {
	// 	fmt.Printf("구독 오류: %v\n", token.Error())
	// } else {
	// 	fmt.Printf("구독자4 %s 가 test/topic 구독 완료\n", subscriberOpts4.ClientID)
	// }

	// 클라이언트가 비정상적으로 종료되었을때 발행되는 will메시지 테스트 -> 비정상적으로 종료되는 경우 시도 실패,, 추후 개발 필요할듯
	publisherOpts3 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1885").
		SetClientID("publisher3").
		SetCleanSession(true).
		// SetUsername("username3").
		// SetPassword("password3").
		SetWill("test/topic3", "will msg test", 1, true) // will 메시지 테스트

	// 세 번째 발행자 클라이언트 생성
	publisherClient3 := mqtt.NewClient(publisherOpts3)
	if token := publisherClient3.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자 클라이언트 3 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자3 %s 가 브로커 [%s]에 연결됨\n", publisherOpts3.ClientID, publisherOpts3.Servers[0].String())

	// 구독자5 클라이언트 생성
	subscriberOpts5 := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1885").
		SetClientID("subscriber5").
		SetDefaultPublishHandler(msgHandler3).
		SetCleanSession(true)

	subscriberClient5 := mqtt.NewClient(subscriberOpts5)
	if token := subscriberClient5.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("구독자 클라이언트 5 브로커에 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("구독자5 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts5.ClientID, subscriberOpts5.Servers[0].String())

	// 구독자5 test/topic3 구독
	if token := subscriberClient5.Subscribe("test/topic3", 0, msgHandler3); token.Wait() && token.Error() != nil {
		fmt.Printf("구독 오류: %v\n", token.Error())
	} else {
		fmt.Printf("구독자5 %s 가 test/topic3 구독 완료\n", subscriberOpts5.ClientID)
	}

	// 발행자 클라이언트 3가 메시지 발행
	message5 := "pub3_msg1"
	token5 := publisherClient3.Publish("test/topic3", 0, true, message5)
	token5.Wait()
	fmt.Printf("발행자3 %s 가 메시지1 발행: %s\n", publisherOpts3.ClientID, message5)
	time.Sleep(3 * time.Second) // 메시지 전송 후 대기

	// 클라이언트 종료
	subscriberClient.Disconnect(250)
	subscriberClient2.Disconnect(250)
	subscriberClient3.Disconnect(250)
	//subscriberClient4.Disconnect(250)
	subscriberClient5.Disconnect(250)
	publisherClient1.Disconnect(250)
	publisherClient2.Disconnect(250)
	publisherClient3.Disconnect(250)

	// // 발행자 클라이언트를 비정상 종료하여 Will 메시지가 발행되도록 테스트
	// publisherClient3.Disconnect(250) // 클라이언트를 정상 종료시키지 않음으로써 Will 메시지 발행 유도

	// // 3초 동안 대기하여 Will 메시지가 발행되고 수신될 시간을 확보
	// time.Sleep(3 * time.Second)

	// // 구독자 클라이언트 종료
	// subscriberClient5.Disconnect(250)

	// // 클라이언트 종료 (비정상 종료) -> 얘로는 will msg 안 나타난다
	// fmt.Println("비정상 종료 시도 중...")
	// os.Exit(1)

	// for { -> 클라이언트 무한 대기하게 한다 -> 무한 대기 중에 netstat -ano | findstr :1885 하고, taskkill /PID 13144 /F 하면 프로세스 강제 종료된다
	// 	// Infinite loop to keep the program running
	// 	time.Sleep(1 * time.Second)
	// }
}

/* retain 메시지

### Retain 메시지의 작동 원리

1. **메시지 발행 시 retain 플래그**: MQTT 클라이언트가 특정 주제에 메시지를 발행할 때, `retain` 플래그를 설정하면, 이 메시지는 브로커에 의해 **저장**된다. 이 저장된 메시지는 주제에 대해 가장 최근에 발행된 retain 메시지가 된다.

2. **브로커가 retain 메시지를 저장**: 브로커는 retain된 메시지를 주제별로 저장한다. 즉, 각 주제마다 마지막에 발행된 retain 메시지를 기억하고 있다. 브로커는 이 메시지를 저장하고 관리하며, 클라이언트나 브로커가 종료되더라도 이 메시지는 유지된다.

3. **새로운 구독자에게 retain 메시지 전달**: 새로운 클라이언트가 특정 주제를 구독하면, 브로커는 해당 주제에 retain된 메시지가 있는지 확인한다. 만약 retain된 메시지가 존재한다면, 브로커는 이 메시지를 즉시 새 구독자에게 전송한다. 이를 통해 새로운 구독자도 주제의 상태를 파악할 수 있게 된다.

### 브로커가 종료된 이후에도 retain 메시지가 유지되는가?
- **브로커의 영속성 설정에 따라 달라짐**: retain 메시지가 브로커의 재시작 후에도 유지되는지는 브로커의 설정에 따라 다르다. 대부분의 MQTT 브로커는 retain 메시지를 디스크에 영구 저장하도록 설정할 수 있다. 이 경우, 브로커가 재시작하더라도 retain된 메시지가 보존된다.
- **기본 설정**: 일반적으로 모스키토(Mosquitto)나 EMQX와 같은 MQTT 브로커들은 기본적으로 retain 메시지를 디스크에 저장하여, 재시작 이후에도 메시지를 잃지 않도록 설정할 수 있다.

### 브로커의 상태와 retain 메시지
- **브로커 재시작 시 유지되는 경우**: 브로커가 retain 메시지를 디스크에 저장한다면, 브로커가 재시작하더라도 retain 메시지가 유지된다. 이는 시스템 종료나 장애가 발생해도 최신 메시지를 새로운 구독자에게 제공할 수 있도록 해준다.
- **브로커의 설정에 따라 다름**: 만약 브로커가 retain 메시지를 메모리에만 저장하도록 설정되어 있다면, 브로커가 재시작되면 이 메시지는 사라질 수 있다. 따라서 디스크 저장을 통해 데이터를 영속적으로 유지하는 것이 중요하다.

### 요약
- **retain 메시지는 브로커가 저장하며, 새로운 구독자에게 전달하기 위해 사용된다.**
- **브로커가 retain 메시지를 디스크에 저장하면, 브로커 종료 후에도 메시지가 유지된다.**
- **브로커의 설정에 따라 retain 메시지의 저장 위치와 지속성은 달라질 수 있다.**

이렇게 retain 메시지는 브로커가 상태를 유지하면서 새로운 클라이언트들에게 주제의 최신 상태를 전달하기 위한 유용한 기능이다.
*/
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

/*
시도해볼 테스트
- qos level 다르게 (ok)
- retain msg (ok)
- will msg (fail)
- subscriber가 수신한 메시지 목록 출력 -> 구독자가 수신한 메시지를 저장할 수 있는 별도의 데이터 구조(예: 배열, 리스트)를 사용 + 뮤텍스 => 메시지를 수신할 때마다 해당 메시지를 리스트에 저장하고, 나중에 이 리스트를 출력
- publisher가 발송한 메시지 목록 출력 -> 발행자가 보낸 메시지를 저장할 수 있는 별도의 데이터 구조(예: 배열 또는 리스트)를 사용 + 뮤텍스 => 발행자가 발송한 모든 메시지를 저장하고 나중에 출력
- subscriber가 구독한 주제 목록 출력 -> MQTT 프로토콜에서는 클라이언트가 구독한 주제 목록을 직접적으로 제공하는 기능이 없다. (보안과 프라이버시 문제) => 구독할 때 주제 정보를 데이터 구조에 저장해두고, 그 정보를 나중에 출력
- 메시지 핸들러에서 구독자id와 토픽, 메시지 출력 -> 클라이언트 객체가 올바르게 초기화된 이후에 ClientID 정보를 가져와야 함
*/
