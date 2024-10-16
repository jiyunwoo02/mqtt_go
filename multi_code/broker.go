package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func startBroker(port string, brokerID string) {
	server := mqtt.New(nil)

	_ = server.AddHook(new(auth.AllowHook), nil)

	tcp := listeners.NewTCP(listeners.Config{
		ID:      brokerID,
		Address: port,
	})

	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		server.Close()
		done <- true
	}()

	<-done
}

func main() {
	// Package flag implements command-line flag parsing.

	// 첫번째 인자는 옵션명(예: -port), 두번째 인자는 기본값, 세번째 인자는 인자에 대한 설명
	// 사용자가 명령행에서 -port나 -id를 입력하지 않으면, 기본값이 사용된다!
	port := flag.String("port", ":1883", "The port on which the broker should run")
	brokerID := flag.String("id", "broker1", "The ID of the broker")

	// 명령행 인자를 파싱하여 정의된 플래그에 값을 할당
	flag.Parse()

	// 플래그 값 출력
	// 각각의 플래그는 포인터 변수에 저장되며, 해당 값을 출력하려면 역참조(*)를 사용
	fmt.Printf("Port: %s in address: %p\n", *port, port) // %p는 포인터의 메모리 주소를 16진수 형식으로 출력
	fmt.Printf("Broker ID: %s in address: %p\n", *brokerID, brokerID)

	// 남은 인자 출력: 플래그로 처리되지 않은 인자, 명령행에 전달되었지만 플래그 이름으로 매칭되지 않은 값
	// 플래그가 아닌 인자(예: 1884)가 등장하면, 그 이후에 있는 모든 값은 남은 인자로 간주!
	fmt.Printf("Remaining args: %v\n", flag.Args()) // %v는 값(value)을 기본 형식으로 출력

	fmt.Printf("Starting broker with ID: %s on port: %s\n\n", *brokerID, *port)

	// If you're using the flags themselves, they are all pointers;
	// flag 패키지가 리턴하는 값은 포인터 -> 값을 출력하려면 역참조해서 값을 가져오도록 해야한다!

	// 역참조 시 port와 brokerID는 string 타입
	startBroker(*port, *brokerID)
}

/*
Q. 포트 번호 앞에 : 은 왜 붙여야 하는가?
A. 포트 번호 앞에 : 를 붙이는 이유는 IP 주소와 포트 번호를 구분하기 위한 표준적인 방식이다.

-> IP 주소와 포트 번호를 함께 사용할 때는 IP주소:포트번호 형식으로 표현
=> 127.0.0.1:1883 은 "IP 주소 127.0.0.1의 1883번 포트에 접근하겠다"를 의미

만약, :1883 처럼 IP 주소 없이 포트 번호만 지정할 때는 :를 앞에 붙여 사용
-> 이는 IP 주소를 생략하고 포트 번호만 지정한 것임을 나타냄
-> 이 경우, 0.0.0.0 또는 모든 네트워크 인터페이스를 기본으로 사용

- IP 주소: 컴퓨터나 네트워크 장치를 식별하는 주소, 예를 들어 127.0.0.1 또는 localhost는 로컬 IP 주소를 나타낸다.
- 포트 번호: 네트워크 상에서 특정 애플리케이션이나 서비스를 식별하는 숫자, 여러 서비스가 동일한 IP 주소에서 동작할 수 있으므로 -> 포트 번호를 사용하여 특정 서비스에 접근. (1883은 MQTT 브로커의 기본 포트 번호)

"1884"처럼 포트 번호만 제공할 때?
- Go의 리스너(net 패키지)가 내부적으로 이를 0.0.0.0:1884로 해석
// Go의 네트워크 라이브러리가 IP 주소를 생략하면 0.0.0.0을 기본값으로 사용
// "1884"처럼 포트 번호만 제공할 때, Go의 리스너(net 패키지)가 내부적으로 이를 0.0.0.0:1884로 해석

1) 0.0.0.0은 모든 네트워크 인터페이스에서 연결을 수락하겠다는 의미
- 서버가 여러 네트워크 인터페이스(예: LAN, Wi-Fi, Ethernet)에서 동시에 연결을 허용하고자 할 때 사용

2) localhost는 루프백(Loopback) 주소를 가리키며, 컴퓨터 자신과의 통신을 위해 사용
- 이는 네트워크 인터페이스가 아닌 소프트웨어적으로 내부 통신을 수행하는 특별한 주소
-- IP 주소로는 127.0.0.1이 사용된다
- 외부 네트워크나 다른 컴퓨터에서는 접근할 수 없으며, 로컬 머신 내의 프로그램들끼리 통신할 때 사용된다.

따라서, 외부 네트워크나 다른 장치에서도 접근 가능한 MQTT 브로커를 설정하려면 0.0.0.0:1883로 리스닝
반면, 로컬 개발 테스트용으로만 사용한다면 **localhost:1883**을 사용
*/
