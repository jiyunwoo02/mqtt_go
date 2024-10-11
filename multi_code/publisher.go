package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run publisher.go <broker_address> <client_id> <topic>")
		return
	}

	// os.Args[0]는 publisher.go
	brokerAddress := os.Args[1]
	clientID := os.Args[2]
	topic := os.Args[3]

	publisherOpts := mqtt.NewClientOptions().
		AddBroker(brokerAddress).
		SetClientID(clientID).
		SetCleanSession(true)

	publisherClient := mqtt.NewClient(publisherOpts)
	if token := publisherClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("발행자1 브로커 연결 실패: %v\n", token.Error())
	}
	fmt.Printf("발행자1 %s 가 브로커 [%s]에 연결됨\n", publisherOpts.ClientID, publisherOpts.Servers[0].String())

	// Enter a loop to get user input and publish messages
	// 사용자가 입력을 완료하고 엔터 키를 누르면 해당 입력을 한 줄로 받아들인다.
	scanner := bufio.NewScanner(os.Stdin) // bufio.Scanner는 줄바꿈(Newline)을 기준으로 입력을 구분
	fmt.Printf("Enter messages to publish to topic '%s' (type 'exit' to quit): ", topic)
	for scanner.Scan() {
		message := scanner.Text()
		if strings.ToLower(message) == "exit" { // 사용자 입력을 소문자로 변환하여 비교
			break
		}

		// 추가 테스트) 메시지의 앞에 'r'이 있는 경우 retain = true
		retain := false
		// HasPrefix reports whether the string s begins with prefix("r").
		if strings.HasPrefix(message, "r") {
			retain = true
			// TrimPrefix returns s without the provided leading prefix string.
			// If s doesn't start with prefix, s is returned unchanged.
			message = strings.TrimPrefix(message, "r") // 'r'을 제거하고 메시지를 발행
			// TrimSpace returns a slice of the string s, with all leading and trailing white space removed, as defined by Unicode
			message = strings.TrimSpace(message) // 앞뒤 공백 제거
		}

		token := publisherClient.Publish(topic, 0, retain, message) // 메시지를 지정된 주제(topic)로 발행
		token.Wait()                                                // 발행이 완료될 때까지 기다리도록 함
		fmt.Printf("Published message: %s\n", message)
	}

	publisherClient.Disconnect(250)
	fmt.Println("Publisher disconnected.")
}

/*
1. scanner := bufio.NewScanner(os.Stdin)
=> bufio.NewScanner를 사용하여 표준 입력(os.Stdin)으로부터 데이터를 읽는 스캐너를 생성
- 이 스캐너는 사용자로부터 입력을 줄 단위로 받을 수 있다.

2. fmt.Printf("Enter messages to publish to topic '%s' (type 'exit' to quit): ", topic)
=> 사용자가 메시지를 입력하면, 해당 메시지가 지정된 주제(topic)로 발행

3. for scanner.Scan() {
    message := scanner.Text()
	...
}
=> 사용자로부터 입력을 받는 무한 루프를 시작
- scanner.Scan()은 사용자로부터 줄 단위 입력을 대기하며, 사용자가 엔터를 누를 때까지 블록된다.
- 사용자가 입력한 내용을 scanner.Text()로 가져와 message 변수에 저장한다.

4. if strings.ToLower(message) == "exit" {
    break
}
=> 사용자가 'exit'라고 입력하면 루프를 빠져나가도록 처리한다.
- strings.ToLower()를 사용하여 대소문자 구분 없이 'exit'을 입력했을 때에도 종료되도록 한다.
*/
