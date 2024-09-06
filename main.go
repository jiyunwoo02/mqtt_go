// main.go
// Eclipse Paho MQTT 클라이언트 라이브러리를 사용하여 로컬 MQTT 브로커에 연결하고, 연결 상태를 확인한 후 연결을 종료
package main

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := "tcp://localhost:1883" // MQTT 브로커 주소는 MQTT 서버의 네트워크 위치 지정 (로컬호스트의 1883 포트 사용)
	clientId := "goClient"           // 클라이언트 ID는 MQTT 네트워크 내에서 각 MQTT 클라이언트를 구분하는 고유 식별자

	opts := mqtt.NewClientOptions() // MQTT 클라이언트 옵션을 생성
	opts.AddBroker(broker)          // 생성된 옵션에 브로커 주소를 추가
	opts.SetClientID(clientId)      // 옵션에 클라이언트 ID를 설정

	client := mqtt.NewClient(opts) // 옵션을 사용하여 새 MQTT 클라이언트 인스턴스 생성
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error()) // 연결 시도 후, 에러가 있다면 에러 출력
		os.Exit(1)                 // 에러가 있을 경우, 프로그램을 비정상 종료
	}

	fmt.Println("Connected to MQTT broker") // 연결이 성공적으로 이루어졌다면, 연결 성공 메시지를 출력
	client.Disconnect(250)                  // 서버와의 연결을 종료하기 전에 250 밀리초 동안 대기
}
