package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func startBroker(port string, brokerID string) {
	// 새로운 MQTT 서버 생성
	server := mqtt.New(nil)

	// 모든 클라이언트의 연결 요청을 허용하도록 인증 훅 추가
	_ = server.AddHook(new(auth.AllowHook), nil)

	// 포트에서 TCP 리스너를 생성
	tcp := listeners.NewTCP(listeners.Config{
		ID:      brokerID,
		Address: port, // 각 브로커가 다른 포트를 사용하도록 설정
	})

	// 서버에 TCP 리스너 추가
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	// 서버 실행
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// 종료 신호 대기
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		server.Close() // 서버 정리
		done <- true
	}()

	<-done // 프로그램이 종료될 때까지 대기
}

func main() {
	// 두 개의 브로커를 각각 포트 1883과 1884에서 실행
	go startBroker(":1883", "broker1")
	go startBroker(":1884", "broker2")
	go startBroker(":1885", "broker3")

	// 메인 함수가 종료되지 않도록 대기
	select {}
}

/*
이 코드는 두 개의 MQTT 브로커를 각기 다른 포트에서 동시에 실행하는 구조를 가지고 있으며, 이를 위해 고루틴을 활용하고 있다.
고루틴과 브로커에 집중해보자.

# 고루틴과 브로커 실행
코드에서 `startBroker` 함수는 새로운 MQTT 브로커를 생성하고,
각 브로커가 지정된 포트에서 클라이언트의 연결을 기다리는 역할을 한다.

고루틴은 이 브로커들이 동시에 실행될 수 있도록 병렬 처리를 담당한다.

1. 고루틴 사용:
   - 메인 함수에서는 `startBroker` 함수를 두 번 호출하는데, 각각의 호출을 `go` 키워드로 고루틴으로 실행
   - `go startBroker(":1883", "broker1")` 와 `go startBroker(":1884", "broker2")` 는 각각 포트 1883과 1884에서 브로커를 실행하는 고루틴을 시작
   - 고루틴은 Go 언어에서 병렬 처리를 가능하게 하며, 이 경우 두 개의 브로커가 동시에 실행되도록 함

2. 브로커 실행:
   - `mqtt.New(nil)`은 새로운 MQTT 서버(브로커)를 생성하고, `listeners.NewTCP` 함수를 통해 TCP 리스너를 특정 포트에서 실행하도록 설정
   - 각각의 브로커는 다른 포트에서 실행되기 때문에 충돌이 발생하지 않는다!
   - `server.Serve()`는 브로커가 클라이언트의 연결 요청을 기다리며 실행되게 만든다

# 브로커 종료 처리
브로커는 운영 체제의 종료 신호
 (예: Ctrl + C)를 감지한 후 안전하게 종료됨

- `signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)` 부분에서 운영 체제 신호를 대기하고, 신호가 수신되면 브로커를 종료하는 `server.Close()`가 호출됨
- `done <- true`는 프로그램이 종료될 때까지 메인 함수가 대기하도록 하는 장치

# 요약
이 코드는 두 개의 MQTT 브로커를 고루틴을 통해 병렬로 실행하고,
각각 다른 포트에서 클라이언트의 연결을 대기하는 구조이다.
운영 체제의 종료 신호를 감지해 브로커를 안전하게 종료한다.

고루틴을 사용해 두 브로커가 동시에 동작하며 각 브로커는 독립적으로 클라이언트와 통신할 수 있다.

추가) 훅이 필수인가?
: 훅(Hook)을 추가하는 것은 MQTT 브로커를 설정할 때 반드시 필요한 것은 아니다.
다만, 훅을 사용하는 것은 브로커의 기능을 확장하거나 특정 동작을 제어하고 싶을 때 매우 유용하다.

-> 훅은 인증, 권한 부여, 메시지 필터링, 로깅, 또는 다른 이벤트 기반 작업을 추가적으로 처리하고자 한다면 훅을 설정하여 해당 기능을 구현할 수 있다.

질문) _ = server.AddHook(new(auth.AllowHook), nil) 없이 실행하면?
: 해당 코드는 Mochi MQTT 브로커에서 모든 클라이언트 연결을 허용하도록 설정하는 기본적인 훅을 추가하는 역할을 한다.
이 코드가 없으면, 기본적으로 클라이언트가 브로커에 연결될 때 인증이 필요하며, 인증되지 않은 클라이언트의 연결이 거부될 수 있다.

-> 이 훅을 추가하지 않으면, MQTT 브로커는 기본적으로 클라이언트의 연결 요청을 거부할 수 있다.
즉, 모든 연결을 수락하도록 설정하지 않으면, MQTT 브로커가 기본적으로 인증 절차를 요구할 수 있어서, 클라이언트 연결이 실패하게 된다.

=> 따라서, 모든 클라이언트 연결을 허용하려면 반드시 이 훅을 추가해야 한다.
그렇지 않으면 클라이언트가 브로커에 연결을 시도할 때 인증 오류가 발생할 수 있으며, 이는 프로그램이 정상적으로 작동하지 않게 되는 원인이 된다.
*/
