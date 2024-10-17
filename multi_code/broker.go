package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func startBroker(port string, brokerID string) {
	server := mqtt.New(nil)

	_ = server.AddHook(new(auth.AllowHook), nil)

	// 포트 번호 앞에 ':'가 없다면 ':'을 붙임
	if !strings.HasPrefix(port, ":") {
		// !는 논리 부정 연산자
		// -> 즉, 포트 번호가 :로 시작하지 않을 때 조건문이 참
		port = ":" + port
	}

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

	// 플래그 정의: 기본값은 ":1883"과 "broker1", 포인터 변수로 저장
	// String defines a string flag with specified name, default value, and usage string.
	// The return value is the address of a string variable that stores the value of the flag.
	// 첫번째 인자는 옵션명(예: -port), 두번째 인자는 기본값(플래그 미제공 시), 세번째 인자는 인자에 대한 설명(--help)
	// -> 사용자가 명령행에서 -port나 -id를 입력하지 않으면, 기본값이 사용된다!
	port := flag.String("port", ":1883", "The port on which the broker should run")
	brokerID := flag.String("id", "broker1", "The ID of the broker")

	// 각각의 플래그는 포인터 변수에 저장되며, 해당 값을 출력하려면 역참조(*)를 사용

	// 파싱 전에 플래그 값 출력 -> 기본값이 해당 메모리 위치에 저장되어 있다.
	// 파싱 전후의 플래그 값은 메모리 주소는 동일 but 그 주소에 저장된 값이 변경될 수 있다!
	fmt.Printf("Before Parsing - Port: %s in address: %p\n", *port, port)
	fmt.Printf("Before Parsing - Broker ID: %s in address: %p\n", *brokerID, brokerID)

	// 파싱 전에 남은 인자 출력 -> 플래그와 남은 인자가 구분되지 않은 상태 -> 남은 인자 정보 제공 X
	// 시도했으나, flag.Args()는 flag.Parse()가 호출된 이후에만 명령행에서 전달된 남은 인자를 반환
	fmt.Printf("Before Parsing - Remaining args: %v\n", flag.Args())

	// 파싱 후 상태 확인: flag.Parse()가 호출된 후에만 true를 반환
	fmt.Printf("Before Parsing - Whether or not it has been parsed: %t\n\n", flag.Parsed())

	// 명령행 인자 파싱 -> 해당 플래그 이름과 일치하는 변수에 값을 저장
	// -> 반드시 flag가 정의가 된 후, flag 값에 접근하기 전 호출
	// 예) -port="1884" → 플래그 port의 값은 "1884"로 설정
	flag.Parse()
	fmt.Println("Parsing Complete\n")

	// 파싱 후 상태 확인: flag.Parse()가 호출된 후에만 true를 반환
	fmt.Printf("After Parsing - Whether or not it has been parsed: %t\n", flag.Parsed())

	// 파싱 후에 플래그 값 출력 (파싱된 값 출력)
	// 사용자가 입력한 명령행 인자들의 값들은 port와 brokerID 포인터가 가리키는 메모리 위치에 덮어씌워진다
	fmt.Printf("After Parsing - Port: %s in address: %p\n", *port, port)
	fmt.Printf("After Parsing - Broker ID: %s in address: %p\n", *brokerID, brokerID)

	// Arg(int index): 명령행 인자 중 플래그로 처리되지 않은 인자들을 순서대로 가져오는 함수, 남은 인자들의 슬라이스에서 해당 인덱스의 요소를 반환
	// Arg(0) is the first remaining argument after flags have been processed.
	// Arg returns an empty string if the requested element does not exist.
	fmt.Printf("After Parsing - First remaining argument: %s\n", flag.Arg(0))
	fmt.Printf("After Parsing - Second remaining argument: %s\n\n", flag.Arg(1))

	// NFlag returns the number of command-line flags that have been set. (입력한 명령행 인자 중 플래그로 처리된 개수)\
	// 사용자가 명령행에서 입력한 인자들 중 플래그 형식(-name=value)으로 정의된 것들만 카운트
	fmt.Printf("After Parsing - Number of flags set: %d\n", flag.NFlag())

	// NArg is the number of arguments remaining after flags have been processed. (남은 인자 수)
	fmt.Printf("After Parsing - Number of remaining arguments: %d\n", flag.NArg())

	// Args returns the non-flag command-line arguments.
	// 남은 인자 출력 (슬라이스로 반환): 플래그로 처리되지 않은 인자, 명령행에 전달되었지만 플래그 이름으로 매칭되지 않은 값
	// 예) 플래그가 아닌 인자 1884 등장 -> 그 이후에 있는 모든 값은 남은 인자로 간주!
	fmt.Printf("After Parsing - Remaining args: %v\n\n", flag.Args())

	// If you're using the flags themselves, they are all pointers;
	// flag 패키지가 리턴하는 값은 포인터 -> 값을 출력하려면 역참조해서 값을 가져오도록 해야한다!

	fmt.Printf("Starting broker with ID: %s on port: %s\n\n", *brokerID, *port)

	// 역참조 시 port와 brokerID는 string 타입
	startBroker(*port, *brokerID)
}

/*

공식 문서: https://pkg.go.dev/flag

Go의 flag 패키지는 명령행 인자에서 플래그 이름에 맞게 변수와 값을 매핑한다.
-> 플래그 이름과 값을 구조체와 포인터로 저장한다.
-> 각 플래그를 FlagSet 구조체 내에 필드로 저장하고, 플래그 값은 포인터를 통해 동적으로 관리

type Flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Value    Value  // value as set -> 플래그 값(string, int..)
	DefValue string // default value (as text); for usage message
}

~ 플래그 정의 및 관리: 모든 플래그가 필드로 저장된다. 플래그 이름과 값(포인터로 관리) 간의 매핑을 관리한다.
type FlagSet struct {
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler. What happens after Usage is called depends
	// on the ErrorHandling setting; for the command line, this defaults
	// to ExitOnError, which exits the program after calling Usage.
	Usage func()
	// contains filtered or unexported fields
	// 내부 필드(비공개)
	// 플래그 이름과 값을 맵을 통해 연결, 사용자가 명령행에서 전달한 값을 맵으로 일시적 저장
}

- 포인터를 사용해 동적으로 값을 관리, 명령행에서 전달된 값이 포인터가 가리키는 메모리에 저장
=> 플래그 이름과 값은 FlagSet 구조체에 저장, 각 플래그는 Flag와 포인터로 참조됨

///

즉, 플래그 이름이 일치하는 변수에 해당 값을 저장한다.
이 과정은 플래그 이름 기반으로 동작하기 때문에 플래그 순서에 상관없이 올바른 값이 변수에 저장된다.
=> 플래그를 사용하려면 -port=1884와 -id=b1처럼 올바른 형식으로 명령행 인자를 전달해야 한다!

# 플래그를 사용해 전달하는 경우: go run broker.go -port="1884" -id="b1"
1. 플래그 정의(코드 실행 준비) -> flag.String() 함수는 포인터(*string: 문자열 변수의 메모리 주소를 가리키는 포인터)를 반환, port 플래그의 기본값은 "1883"이고, brokerID의 기본값은 "broker1"
2. 명령행 인자 파싱 -> flag.Parse()는 명령행에서 전달된 인자들을 파싱하여 해당 플래그 이름과 일치하는 변수에 값을 저장, 플래그 port의 값은 ":1884"로, 플래그 brokerID의 값은 "b1"로 설정
3. 플래그 값 출력 (역참조 사용) -> 역참조(*port)를 사용해 포인터가 가리키는 값을 출력
4. 남은 인자 처리 -> flag.Args()는 플래그로 처리되지 않은 인자들을 반환, 이 경우 모든 명령행 인자가 플래그와 매칭되었으므로 남은 인자 X
5. 브로커 시작 -> 프로그램은 입력된 값에 따라 브로커를 b1 ID와 포트 1884로 시작

# 일반 인자만 전달하는 경우: go run broker.go 1884 b1
: 플래그로 처리되지 않은 인자는 남은 인자(flag.Args())로 취급
1. 플래그 정의 (기본값 설정) -> port 플래그의 기본값: "1883", brokerID 플래그의 기본값: "broker1"
2. 명령행 인자 파싱 -> 플래그 형식(-port="1884")이 아닌 일반 인자인 1884와 b1을 플래그로 인식 X
3. 플래그 값 출력 -> port와 brokerID는 기본값("1883"과 "broker1")을 사용, 명령행 인자 1884와 b1은 남은 인자로 처리
4. 브로커 시작 -> 브로커가 기본값(port="1883", brokerID="broker1")으로 실행

# 하나는 플래그, 하나는 일반 인자를 사용해 전달하는 경우
1) go run broker.go 1884 -id="b1" : 포트와 아이디는 기본 값으로 브로커 구동
: 1884는 플래그 형식이 아니기 때문에 일반 인자로 처리
: -id="b1"은 플래그 형식이 맞지만, 첫 번째 일반 인자(1884)가 등장한 후이므로, 더 이상 플래그로 인식 X

-> Go의 flag 패키지는 일반 인자가 등장하면, 그 이후의 모든 인자를 남은 인자로 처리한다!
=> 따라서 1884가 일반 인자로 등장한 이후, -id="b1"도 플래그로 인식되지 않고 남은 인자로 처리된다.

2) go run broker.go -port="1884" b1 : 포트는 1884로, 아이디는 기본 값으로
: -port="1884"은 올바른 플래그로 인식, port 플래그 값이 1884로 설정
: b1은 플래그가 아니므로 남은 인자로 처리

*/

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
반면, 로컬 개발 테스트용으로만 사용한다면 localhost:1883을 사용
*/

// %p는 포인터의 메모리 주소를 16진수 형식으로 출력
// %v는 값(value)을 기본 형식으로 출력
// %t는 부울 값을 출력할 때 사용하는 형식 지시자
