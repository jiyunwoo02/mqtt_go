package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/config"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/DrmagicE/gmqtt/plugin/auth"
)

// CustomHook는 메시지 수신 시 처리할 훅을 정의합니다.
type CustomHook struct{}

// OnMsgArrived는 메시지 발행 시 호출되는 함수입니다.
func (h *CustomHook) OnMsgArrived(ctx context.Context, client *server.Client, msg *packets.Publish) (valid bool) {
	fmt.Printf("클라이언트 %s가 메시지를 발행했습니다: %s\n", client.ClientOptions().ClientID, string(msg.Payload))
	return true
}

func main() {
	// 브로커 설정을 위한 기본 구성 생성
	cfg, err := config.ParseConfig([]byte(config.DefaultConfig()))
	if err != nil {
		log.Fatalf("브로커 설정 생성 실패: %v", err)
	}

	// 브로커 인스턴스 생성
	broker := server.New(
		server.WithConfig(cfg),
	)

	// 인증 플러그인 추가: 모든 클라이언트를 허용
	authPlugin := auth.NewAllowAll()
	broker.AddPlugin(authPlugin)

	// 브로커에 CustomHook 추가
	hook := &CustomHook{}
	broker.AddHook(hook)

	// TCP 리스너 추가
	err = broker.AddTCPListener(":1883")
	if err != nil {
		log.Fatalf("TCP 리스너 추가 실패: %v", err)
	}

	// 브로커 실행
	go func() {
		if err := broker.Run(); err != nil {
			log.Fatalf("브로커 실행 실패: %v", err)
		}
	}()
	fmt.Println("브로커가 포트 1883에서 실행 중입니다...")

	// 시스템 신호 처리 (종료를 위해 대기)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// 브로커 종료
	broker.Close()
	fmt.Println("브로커가 종료되었습니다.")
}
