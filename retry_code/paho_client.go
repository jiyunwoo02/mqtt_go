// package main

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// // 메시지 수신 핸들러: 주제에 대해 발행한 메시지가 구독자에게 발행된 경우에 실행되는 callback 타입
// // 메시지 핸들러가 설정되지 않으면(nil), MQTT 클라이언트는 특정 메시지를 수신할 때 아무런 동작을 수행하지 않는다!
// var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("수신한 메시지: 토픽 [%s] - 메시지 [%s]\n", msg.Topic(), msg.Payload())
// }

// func main() {
// 	// 발행자 클라이언트 옵션 설정
// 	publisherOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 브로커 주소
// 		SetClientID("publisher").          // 발행자 클라이언트 ID 설정
// 		SetUsername("username").           // 사용자 이름 설정 (필요시)
// 		SetPassword("password")            // 비밀번호 설정 (필요시)

// 	// 발행자 클라이언트 객체 생성
// 	publisherClient := mqtt.NewClient(publisherOpts)
// 	if token := publisherClient.Connect(); token.Wait() && token.Error() != nil { // Token defines the interface for the tokens used to indicate when actions have completed.
// 		log.Fatalf("발행자 클라이언트 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("발행자 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())
// 	// 발행자 publisherClient 가 브로커 [tcp://localhost:1883]에 연결됨
// 	// 발행자 객체의 Servers 필드 -> 첫 번째 브로커의 주소 가져와 (여러 브로커 연결된 경우 대비, 배열로 저장) 문자열로 반환

// 	// 구독자 클라이언트 옵션 설정
// 	subscriberOpts := mqtt.NewClientOptions().
// 		AddBroker("tcp://localhost:1883"). // 브로커 주소
// 		SetClientID("subscriber")          // 구독자 클라이언트 ID 설정

// 	// 구독자 클라이언트 객체 생성
// 	subscriberClient := mqtt.NewClient(subscriberOpts)
// 	if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
// 		log.Fatalf("구독자 브로커에 연결 실패: %v\n", token.Error())
// 	}
// 	fmt.Printf("구독자 %s 가 브로커 [%s]에 연결됨\n", subscriberOpts.ClientID, subscriberOpts.Servers[0].String())

// 	// test/topic 구독 - qos 0,1,2로 바꾸면서 패킷 확인
// 	if token := subscriberClient.Subscribe("test/topic", 1, msgHandler); token.Wait() && token.Error() != nil {
// 		fmt.Printf("구독 오류: %v\n", token.Error())
// 	} else {
// 		fmt.Printf("구독자 %s 가 test/topic 구독 완료\n", subscriberOpts.ClientID)
// 	}

// 	// 발행자 클라이언트가 메시지 발행 - qos 0,1,2로 바꾸면서 패킷 확인
// 	message := "Hello, MQTT!"
// 	token := publisherClient.Publish("test/topic", 1, false, message)
// 	token.Wait()
// 	fmt.Printf("발행자 %s 가 메시지 발행: %s\n", publisherOpts.ClientID, message)

// 	// 3sec 동안 대기하여 메시지 수신 대기
// 	time.Sleep(3 * time.Second)

// 	// 250millisec 기다린 후 클라이언트 종료
// 	subscriberClient.Disconnect(250)
// 	publisherClient.Disconnect(250)

// 	fmt.Println("클라이언트 종료됨")
// }

// /* 부가 설명

// 1. token이란?
// : 클라이언트가 서버와의 비동기 작업(연결, 메시지 발행, 구독 등)의 상태와 결과를 추적하고 관리

// - In the context of the Paho MQTT library,
// 	a token is a structure used to track the state of an MQTT operation,
// 	such as a connection, publish, subscribe, or unsubscribe action.
// - It represents an asynchronous operation and
// 	allows you to track whether the operation has completed,
// 	whether it has succeeded or failed,
// 	and retrieve any error messages if the operation failed.
// - In summary, while token may seem like a data type,
// 	it is more accurately described as a state-tracking mechanism for MQTT operations in the Paho library.

// - Technically, the token in Paho is a struct that implements the Token interface in Go.

// - The Token interface provides methods such as:
// 	1) Wait(): Blocks the execution until the operation represented by the token is complete.
// 	2) WaitTimeout(timeout time.Duration): Blocks the execution for a specified time or until the operation is complete.
// 	3) Error(): Returns the error if the operation failed, otherwise it returns nil.

// -> Since it implements an interface, it's not just a basic data type,
// 	but a mechanism to handle the state of the operation.
// 	It encapsulates the process of checking whether the MQTT operation was successful or encountered any errors.

// 2. 코드 분석
// if token := subscriberClient.Connect(); token.Wait() && token.Error() != nil {
// 	// 1. 객체가 브로커에 연결을 시도 후 결과 나타내는 token 객체 반환
// 	// 2. 연결 완료될 때까지 블로킹하여 대기, 호출 끝나면 연결 시도 끝났음을 의미
// 	// - 성공했는지 실패했는지는 token 객체의 상태 통해 알 수 있다!
// 	// 3. 연결 시도 중 발생한 에러가 nil이 아니라면 연결 시도 실패했음을 의미
// 	log.Fatalf("구독자 브로커에 연결 실패: %v\n", token.Error())
// 	// subscriberClient가 MQTT 브로커에 연결을 시도한 후,
// 	만약 연결이 실패하면 프로그램을 종료하고 "구독자 브로커에 연결 실패"라는 에러 메시지와 함께 발생한 에러 내용을 출력
// }

// 3. token의 주요 기능
// 1) 작업 상태 확인:
// 	token은 클라이언트와 서버 간의 작업(연결, 메시지 발행 등)이 성공적으로 완료되었는지, 아니면 오류가 발생했는지를 확인.

// 2) Wait() 메서드:
// 	token.Wait()는 해당 작업이 완료될 때까지 기다리는 메서드.
// 	예를 들어, 브로커에 연결이 완료될 때까지 대기하거나 메시지가 발행될 때까지 대기.

// 3) Error() 메서드:
// 	token.Error()는 작업이 성공적으로 완료되었는지, 아니면 오류가 발생했는지를 확인.
// 	작업이 성공하면 nil을 반환하고, 오류가 발생하면 오류 내용을 반환.

// 4. Token 인터페이스 - paho.mqtt.golang
// : token 자체는 Token 인터페이스의 구현체

// -> Wait() bool. WaitTimeout(time.Duration) bool, Done() <-chan struct{}, Error() error 메서드 정의

// [구현체]
// - ClientToken은 클라이언트 연결 작업을 추적
// - PublishToken은 메시지 발행 작업을 관리
// - SubscribeToken은 구독 작업을 추적

// 4. callback function
// : 콜백 함수는 전달인자로 다른 함수에 전달되는 함수

// [위키백과]
// 프로그래밍에서 콜백(callback) 또는 콜백 함수(callback function)는 다른 코드의 인수로서 넘겨주는 실행 가능한 코드를 말한다.
// 콜백을 넘겨받는 코드는 이 콜백을 필요에 따라 즉시 실행할 수도 있고, 아니면 나중에 실행할 수도 있다.

// 일반적으로 콜백수신 코드로 콜백 코드(함수)를 전달할 때는 콜백 함수의 포인터 (핸들), 서브루틴 또는 람다함수의 형태로 넘겨준다.
// 콜백수신 코드는 실행하는 동안에 넘겨받은 콜백 코드를 필요에 따라 호출하고 다른 작업을 실행하는 경우도 있다.
// 다른 방식으로는 콜백수신 코드는 넘겨받은 콜백 함수를 '핸들러'로서 등록하고, 콜백수신 함수의 동작 중 어떠한 반응의 일부로서 나중에 호출할 때 사용할 수도 있다 (비동기 콜백).

// 콜백은 코드 재사용을 할 때 유용하다.

// 5. 메시지 핸들러 설정
// if token := subscriberClient.Subscribe("test/topic", 0, msgHandler); token.Wait() && token.Error() != nil

// - MQTT 클라이언트에서 메시지 핸들러(MessageHandler)가 nil로 설정된 경우,
// 	이는 클라이언트가 기본 메시지 처리 함수를 제공하지 않았다는 것을 의미
// 	메시지 핸들러가 설정되지 않으면, MQTT 클라이언트는 특정 메시지를 수신할 때 아무런 동작을 수행하지 않는다.

// 즉, 구독한 주제에서 메시지가 도착해도, 메시지 핸들러가 없기 때문에 그 메시지를 처리할 로직이 없는 상태가 된다.
// 이런 경우에는 기본적인 콜백 함수가 호출되지 않고, 클라이언트는 해당 메시지를 무시하게 된다.

// -> 메시지를 수신하기 위해서는 메시지 핸들러가 반드시 설정되어야 함
// 이를 위해 SetDefaultPublishHandler 메서드를 사용하거나, Subscribe 메서드 호출 시 개별 주제마다 메시지 핸들러를 설정해야 한다

// 추가) SetDefaultPublishHandler 메서드란?
// : Eclipse Paho MQTT Go 라이브러리에서 제공하는 메서드로, MQTT 클라이언트가 수신한 메시지를 처리하기 위한 기본 핸들러를 설정하는 데 사용
// : 클라이언트가 구독하고 있는 주제로부터 메시지를 수신할 때 호출된다
// : 메시지가 수신되었지만, 해당 주제에 대한 별도의 핸들러가 설정되지 않은 경우에 호출된다

// */
