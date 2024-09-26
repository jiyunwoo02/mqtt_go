package main

import (
	"log"       // 로깅을 위한 패키지
	"os"        // 운영 체제와 상호작용하기 위한 패키지
	"os/signal" // 운영 체제의 신호를 처리하기 위한 패키지
	"syscall"   // 시스템 호출 인터페이스를 제공하는 패키지

	mqtt "github.com/mochi-mqtt/server/v2"       // Mochi MQTT 서버를 위한 패키지
	"github.com/mochi-mqtt/server/v2/hooks/auth" // 인증을 처리하는 훅 패키지
	"github.com/mochi-mqtt/server/v2/listeners"  // 리스너 설정을 위한 패키지
)

func main() {
	// 서버가 종료될 때까지 신호를 대기하기 위한 채널 생성
	sigs := make(chan os.Signal, 1) // 신호 수신, buffer = 1
	done := make(chan bool, 1)      // 서버가 종료될 때 사용, buffer = 1

	// SIGINT(인터럽트, Ctrl + C) 또는 SIGTERM(종료)을 받으면 신호 채널에 전달
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 새로운 MQTT 서버 생성
	server := mqtt.New(nil) // 기본 설정 사용하여 서버 생성

	// 모든 연결 요청을 허용하도록 인증 훅 추가
	_ = server.AddHook(new(auth.AllowHook), nil) // nil : 추가 설정 없이 기본 동작을 사용
	// auth.AllowHook은 기본적으로 모든 인증 요청을 허용하는 훅 (=인증 및 권환 확인 절차 pass)

	// 기본 포트(1883)에서 TCP 리스너를 생성
	tcp := listeners.NewTCP(listeners.Config{ // 1883 포트에서 수신할 TCP 리스너를 생성
		ID:      "t1",    // 리스너 ID 설정
		Address: ":1883", // 리스너 주소 및 포트 설정
	})

	// 서버에 TCP 리스너를 추가합니다.
	// -> 서버는 리스너가 수신하는 모든 연결을 처리
	err := server.AddListener(tcp)
	if err != nil { // 리스너 추가 시 오류가 발생하면 로그 출력하고 프로그램 종료
		log.Fatal(err)
	}

	// 신호 수신될 때까지 대기하는 고루틴
	go func() {
		// 화살표(<-) 왼쪽에 아무것도 없는 것은 "채널에서 값을 수신하여 변수에 저장하지 않겠다"는 의미
		<-sigs         // sigs 채널에 데이터 들어올 때까지 대기 -> SIGINT나 SIGTERM 신호가 들어올 때까지 고루틴이 멈춰있다가, 신호가 들어오면 다음 구문으로 이동
		server.Close() // 서버 정리 작업: mochi-mqtt 서버의 모든 리스너를 종료하고, 서버 인스턴스를 안전하게 종료
		done <- true   // 신호가 수신되면 done 채널에 true 보냄 -> 프로그램 안전하게 종료됨
	}()

	// 서버 실행하는 고루틴
	go func() {
		err := server.Serve() // 서버 시작, 클라이언트 연결 수신, 서버가 실행 중인 동안 이 고루틴은 종료되지 않음
		if err != nil {       // 서버 실행 중 오류 발생하면 로그 출력하고 프로그램 종료
			log.Fatal(err)
		}
	}()

	// 프로그램 종료될 때까지 대기
	<-done // 메인 함수는 done 채널에서 값을 받을 때까지 대기
	// done 채널은 main() 함수 위쪽에서 정의된 done <- true에 의해 값이 전달된다

	// 신호를 대기하는 고루틴이 done <- true를 보내면 메인 고루틴에서 <-done에 의해 값이 수신되고 프로그램이 종료된다.
	// -> 이 경우 서버를 실행 중인 고루틴(server.Serve())도 함께 강제 종료됨 => 서버는 비정상적으로 종료될 수 있다. => server.Close()
}

/* 부가 설명

1. TCP Listener란?
: 네트워크 소켓을 열어서 특정 포트(예: 1883)에서 클라이언트 연결을 대기하는 역할
-> 서버가 클라이언트와 통신하기 위해서는 반드시 리스너가 필요 [클라이언트 연결 대기, 연결 수락, 연결 관리 역할]

클라이언트가 서버의 특정 포트에 연결하려고 하면, TCP 리스너가 그 요청을 받아들여 서버의 애플리케이션 로직(예: 메시지 처리, 데이터베이스 작업 등)으로 연결을 넘겨준다!

*/
