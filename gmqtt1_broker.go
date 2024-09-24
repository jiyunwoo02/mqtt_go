package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/persistence"
)

func main() {
	// MQTT 브로커 생성
	server := gmqtt.NewServer(
		gmqtt.WithTCPListener(":1883"),
	)

	// 브로커의 메시지 저장소 설정 (선택 사항)
	server.SetStore(persistence.NewMemoryStore())

	// 클라이언트가 연결될 때마다 호출되는 콜백 함수 설정
	server.AddHook(new(gmqtt.Hook), &gmqtt.HookOptions{
		OnConnected: func(ctx context.Context, client gmqtt.Client) {
			log.Printf("클라이언트 %s 연결됨", client.ClientOptions().ClientID)
		},
	})

	// 브로커 실행
	if err := server.Run(); err != nil {
		log.Fatalf("브로커 실행 실패: %v", err)
	}

	fmt.Println("MQTT 브로커가 1883 포트에서 실행 중입니다...")

	// 브로커가 종료될 때까지 대기
	select {}
}
