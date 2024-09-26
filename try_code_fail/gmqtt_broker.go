package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "try_code/persistence" // memory persistence를 import하여 등록

	"github.com/DrmagicE/gmqtt/config"
	"github.com/DrmagicE/gmqtt/server"
	"gopkg.in/yaml.v2"
)

type BrokerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func LoadConfig(configFile string) (config.Config, *BrokerConfig, error) {
	var cfg config.Config
	var brokerCfg BrokerConfig

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, nil, fmt.Errorf("설정 파일을 읽을 수 없습니다: %v", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, nil, fmt.Errorf("설정 파일 파싱 오류: %v", err)
	}

	if err := yaml.Unmarshal(data, &brokerCfg); err != nil {
		return cfg, nil, fmt.Errorf("브로커 설정 파싱 오류: %v", err)
	}

	return cfg, &brokerCfg, nil
}

func main() {
	cfg, brokerCfg, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("설정 파일 로드 오류: %v\n", err)
		return
	}

	tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", brokerCfg.Address, brokerCfg.Port))
	if err != nil {
		fmt.Printf("TCP 리스너를 생성할 수 없습니다: %v\n", err)
		return
	}

	// 서버 초기화
	srv := server.New(
		server.WithTCPListener(tcpListener),
		server.WithConfig(cfg),
	)

	go func() {
		if err := srv.Run(); err != nil {
			fmt.Printf("서버 실행 중 오류: %v\n", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		fmt.Printf("서버 종료 오류: %v\n", err)
	} else {
		fmt.Println("서버가 정상적으로 종료되었습니다.")
	}
}
