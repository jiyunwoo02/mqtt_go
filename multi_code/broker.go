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

	// Define command-line arguments
	// 첫번째 인자는 옵션명(예: -port), 두번째 인자는 기본값(default), 세번째 인자는 인자가 잘못 넘어왔을때 표현될 설명
	port := flag.String("port", ":1883", "The port on which the broker should run")
	brokerID := flag.String("id", "broker1", "The ID of the broker")

	// Parse the command-line arguments -> into the defined flags.
	flag.Parse()

	// Displaying the parsed arguments
	fmt.Printf("Starting broker with ID: %s on port: %s\n\n", *brokerID, *port)

	// If you're using the flags themselves, they are all pointers;
	// flag 패키지가 리턴하는 값은 포인터 -> 값을 출력하려면 역참조해서 값을 가져오도록 해야한다!

	// Start the broker with the provided port and broker ID
	startBroker(*port, *brokerID)

	// go run broker.go 1884 broker2 를 터미널에 입력 시 -> 포트 1884에서 broker2라는 id로 브로커 구동
}
