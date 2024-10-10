// package main

// import (
// 	"fmt"
// 	"log"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// /* 여기에서 수신자가 누구인지 clientid를 출력을 못할까? -> ClientOptionsReader가 있다
// ClientOptionsReader provides an interface for reading ClientOptions after the client has been initialized.
// = 클라이언트가 초기화된 이후에 사용할 수 있다
// 주석 코드 실행하면 cannot call pointer method ClientID on ~ .ClientOptionsReader
// */

// // 메시지 수신 핸들러
// var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	// cliID := client.OptionsReader().ClientID()
// 	fmt.Printf("주제1에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
// }

// var msgHandler2 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("주제2에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
// }

// var msgHandler3 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("주제3에 대해 수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
// }

// func main() {
// 	// 첫 번째 브로커에 연결하는 발행자1 클라이언트 옵션 설정 (포트 1883)
// 	publisherOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
// 		SetClientID("publisher1").
// 		SetUsername("username1").
// 		SetPassword("password1")

// 	// 첫 번째 발행자 클라이언트 생성
// 	publisherClient1 := mqtt.NewClient(publisherOpts)
// 	if token := publisherClient1.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("발행자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자1 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())

// 	// 두 번째 브로커에 연결하는 발행자2 클라이언트 옵션 설정 (포트 1884)
// 	publisherOpts2 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1884"). // 두 번째 브로커 주소
// 		SetClientID("publisher2").
// 		SetUsername("username2").
// 		SetPassword("password2")

// 	// 두 번째 발행자 클라이언트 생성
// 	publisherClient2 := mqtt.NewClient(publisherOpts2)
// 	if token := publisherClient2.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("발행자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자2 %s 가 브로커 [%s]에 연결됨\n", publisherOpts2.ClientID, publisherOpts2.Servers[0].String())

// 	// 구독자1 클라이언트 옵션 설정 (포트 1883)
// 	subscriberOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 첫 번째 브로커 주소
// 		SetClientID("subscriber1").
// 		SetDefaultPublishHandler(msgHandler)

// 	// 구독자1 클라이언트 생성
// 	subscriberClient := mqtt.NewClient(subscriberOpts)
// 	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 1 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자1 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

// 	// 구독자2 클라이언트 옵션 설정 (포트 1884)
// 	subscriberOpts2 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1884").
// 		SetClientID("subscriber2").
// 		SetDefaultPublishHandler(msgHandler2)

// 	// 구독자2 클라이언트 생성
// 	subscriberClient2 := mqtt.NewClient(subscriberOpts2)
// 	if token := subscriberClient2.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 2 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자2 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts2.ClientID, subscriberOpts2.Servers[0].String())

// 	// 구독자1 test/topic 구독
// 	if token := subscriberClient.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자1 %s 가 test/topic 구독 완료\n", subscriberOpts.ClientID)
// 	}

// 	// 구독자2 test/topic2 구독
// 	if token := subscriberClient2.Subscribe("test/topic2", 0, msgHandler2); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자2 %s 가 test/topic2 구독 완료\n", subscriberOpts2.ClientID)
// 	}

// 	// 발행자 클라이언트 1이 메시지 발행
// 	message := "Hello from Publisher1!"
// 	token := publisherClient1.Publish("test/topic", 0, false, message)
// 	token.Wait()
// 	fmt.Printf("발행자1 %s 가 메시지 발행: %s\n", publisherOpts.ClientID, message)

// 	// 발행자 클라이언트 2가 메시지 발행
// 	message2 := "Hello from Publisher2!"
// 	token2 := publisherClient2.Publish("test/topic2", 0, false, message2)
// 	token2.Wait()
// 	fmt.Printf("발행자2 %s 가 메시지 발행: %s\n", publisherOpts2.ClientID, message2)

// 	// 이미 주제에 대해 첫 메시지 발행한 이후!
// 	// 구독자3 클라이언트 옵션 설정 (포트 1884)
// 	subscriberOpts3 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1884").
// 		SetClientID("subscriber3").
// 		SetDefaultPublishHandler(msgHandler2)

// 	// 구독자3 클라이언트 생성
// 	subscriberClient3 := mqtt.NewClient(subscriberOpts3)
// 	if token := subscriberClient3.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 3 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자3 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts3.ClientID, subscriberOpts3.Servers[0].String())

// 	// 구독자3 test/topic2 구독
// 	if token := subscriberClient3.Subscribe("test/topic2", 0, msgHandler2); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자3 %s 가 test/topic2 구독 완료\n", subscriberOpts3.ClientID)
// 	}

// 	// 발행자 클라이언트 2가 메시지 발행
// 	message3 := "Hello2 from Publisher2!"
// 	token3 := publisherClient2.Publish("test/topic2", 0, false, message3)
// 	token3.Wait()
// 	fmt.Printf("발행자2 %s 가 메시지 발행: %s\n", publisherOpts2.ClientID, message3)

// 	// 발행자 클라이언트 1가 메시지 발행
// 	message4 := "Hello2 from Publisher1!"
// 	token4 := publisherClient1.Publish("test/topic", 0, true, message4)
// 	token4.Wait()
// 	fmt.Printf("발행자1 %s 가 메시지 발행 (retain): %s\n", publisherOpts.ClientID, message4)

// 	// 구독자4가 주제가 이미 메시지를 발행한 후에 구독 -> retain=true를 통해 전달받음을 확인
// 	// 구독자4 클라이언트 옵션 설정 (포트 1883)
// 	subscriberOpts4 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883").
// 		SetClientID("subscriber4")

// 	// 구독자4 클라이언트 생성
// 	subscriberClient4 := mqtt.NewClient(subscriberOpts4)
// 	if token := subscriberClient4.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 4 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자4 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts4.ClientID, subscriberOpts4.Servers[0].String())

// 	// 구독자4 test/topic 구독
// 	if token := subscriberClient4.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자4 %s 가 test/topic 구독 완료\n", subscriberOpts4.ClientID)
// 	}

// 	publisherOpts3 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1885").
// 		SetClientID("publisher3").
// 		SetUsername("username3").
// 		SetPassword("password3").
// 		SetWill("test/topic3", "will msg test", 1, true) // will 메시지 테스트

// 	// 세 번째 발행자 클라이언트 생성
// 	publisherClient3 := mqtt.NewClient(publisherOpts3)
// 	if token := publisherClient3.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("발행자 클라이언트 3 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자3 %s 가 브로커 [%s]에 연결됨\n", publisherOpts3.ClientID, publisherOpts3.Servers[0].String())

// 	// 구독자5 클라이언트 생성
// 	subscriberOpts5 := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1885").
// 		SetClientID("subscriber5").
// 		SetDefaultPublishHandler(msgHandler3)

// 	subscriberClient5 := mqtt.NewClient(subscriberOpts5)
// 	if token := subscriberClient5.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 클라이언트 5 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자5 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts5.ClientID, subscriberOpts5.Servers[0].String())

// 	// 구독자5 test/topic3 구독
// 	if token := subscriberClient5.Subscribe("test/topic3", 0, msgHandler3); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자5 %s 가 test/topic3 구독 완료\n", subscriberOpts5.ClientID)
// 	}

// 	// 발행자 클라이언트 1가 메시지 발행
// 	message5 := "Hello3 from Publisher3!"
// 	token5 := publisherClient3.Publish("test/topic3", 0, true, message5)
// 	token5.Wait()
// 	fmt.Printf("발행자3 %s 가 메시지 발행: %s\n", publisherOpts3.ClientID, message5)

// 	// 클라이언트 종료
// 	subscriberClient.Disconnect(250)
// 	subscriberClient2.Disconnect(250)
// 	subscriberClient3.Disconnect(250)
// 	subscriberClient4.Disconnect(250)
// 	subscriberClient5.Disconnect(250)
// 	publisherClient1.Disconnect(250)
// 	publisherClient2.Disconnect(250)
// 	publisherClient3.Disconnect(250)

// 	// // 발행자 클라이언트를 비정상 종료하여 Will 메시지가 발행되도록 테스트
// 	// publisherClient3.Disconnect(250) // 클라이언트를 정상 종료시키지 않음으로써 Will 메시지 발행 유도

// 	// // 3초 동안 대기하여 Will 메시지가 발행되고 수신될 시간을 확보
// 	// time.Sleep(3 * time.Second)

// 	// // 구독자 클라이언트 종료
// 	// subscriberClient5.Disconnect(250)

// 	// // 클라이언트 종료 (비정상 종료) -> 얘로는 will msg 안 나타난다
// 	// fmt.Println("비정상 종료 시도 중...")
// 	// os.Exit(1)

// 	// for { -> 클라이언트 무한 대기하게 한다 -> 무한 대기 중에 netstat -ano | findstr :1885 하고, taskkill /PID 13144 /F 하면 프로세스 강제 종료된다
// 	// 	// Infinite loop to keep the program running
// 	// 	time.Sleep(1 * time.Second)
// 	// }
// }

// /*
// 동일한 주제를 구독하고 있더라도 서로 다른 브로커에 연결된 클라이언트들은 서로의 메시지를 수신하지 못한다.
// 클라이언트가 메시지를 받기 위해서는 발행자와 구독자가 같은 브로커에 연결되어 있어야 한다.

// -> 이를 해결하기 위해서는 브로커 간의 브릿지(bridge) 설정을 통해 브로커 간 메시지를 공유하도록 설정할 수도 있다.
// 하지만 이는 기본 설정에서는 제공되지 않으며, 추가적인 설정이 필요하다.
// */

// /* 주제를 구독한 구독자들의 목록은 출력하지 못할까? -> 보안상의 문제로 못하는듯
// In the MQTT protocol, there isn't a built-in mechanism for clients
// 	to directly obtain a list of other clients subscribed to a specific topic.

// MQTT brokers generally do not provide this information to clients due to privacy and security concerns.
// */

// /*
// 시도해볼 테스트
// - qos level 다르게 (ok)
// - retain msg (ok)
// - will msg (fail)
// - subscriber가 수신한 메시지 목록 출력 -> 구독자가 수신한 메시지를 저장할 수 있는 별도의 데이터 구조(예: 배열, 리스트)를 사용 + 뮤텍스 => 메시지를 수신할 때마다 해당 메시지를 리스트에 저장하고, 나중에 이 리스트를 출력
// - publisher가 발송한 메시지 목록 출력 -> 발행자가 보낸 메시지를 저장할 수 있는 별도의 데이터 구조(예: 배열 또는 리스트)를 사용 + 뮤텍스 => 발행자가 발송한 모든 메시지를 저장하고 나중에 출력
// - subscriber가 구독한 주제 목록 출력 -> MQTT 프로토콜에서는 클라이언트가 구독한 주제 목록을 직접적으로 제공하는 기능이 없다. (보안과 프라이버시 문제) => 구독할 때 주제 정보를 데이터 구조에 저장해두고, 그 정보를 나중에 출력
// - 메시지 핸들러에서 구독자id와 토픽, 메시지 출력 -> 클라이언트 객체가 올바르게 초기화된 이후에 ClientID 정보를 가져와야 함
// */
