# mqtt_go
MQTT 프로토콜 스터디 진행 - GoLang
- https://velog.io/@jiyunwoo/series/MQTTGo 참조

## 1. MQTT란?
MQTT(Message Queuing Telemetry Transport)
: 경량의 메시지 프로토콜, 낮은 대역폭이나 불안정한 네트워크 환경에서도 장치 간에 메시지를 효율적으로 전달할 수 있게 설계됨
(1999년에 IBM과 유럽의 한 유틸리티 회사가 개발, IoT 장치와 서버 간의 통신을 위해 널리 사용됨)

- MQTT는 퍼블리시/구독(Publish/Subscribe) 모델을 사용
- 메시지는 주제(Topic)별로 분류
- 구독자만 그 주제의 메시지를 받을 수 있다

이 모델은 중앙 서버(브로커)를 통해 작동
클라이언트는 브로커에게 메시지를 보내거나, 브로커로부터 메시지를 받는다
-> 이를 통해 다수의 수신자에게 메시지를 효율적으로 배포할 수 있다

[주요 특징]
1. 경량 프로토콜: 매우 적은 데이터 양을 사용하여 통신 가능
2. 저전력: 배터리로 구동되는 장치에 적합하도록 설계됨
3. 높은 신뢰성: 메시지의 QoS(Quality of Service) 수준을 설정하여 메시지의 신뢰성과 전달 여부 제어 가능
4. 보안: TLS/SSL을 통한 메시지 암호화 지원

특히 사물인터넷 환경에서 다양한 장치와 플랫폼 간의 연결성과 데이터 전송을 간소화하고자 할 때 매우 유용한 프로토콜

## 2. 디렉토리 소개
- gmqtt_code_fail : GoLang의 gmqtt 라이브러리를 사용해 브로커/클라이언트 구동 시도 -> 실패 (추후 재시도)
- mochi_code_succeed : GoLang의 mochi 라이브러리를 사용해 브로커 구동 시도 -> 성공
- paho_code_succeed : GoLang의 paho 라이브러리를 사용해 브로커(mosquitto) 및 클라이언트 구동 및 연동 -> 성공
- try_code_fail : GoLang의 gmqtt 라이브러를 사용해 브로커를, paho를 사용해 클라이언트 구동 및 연동 시도 -> gmqtt 실패, paho 성공 (추후 재시도)
- retry_code : GoLang의 mochi 라이브러를 사용해 브로커를, paho를 사용해 클라이언트 구동 및 연동 시도 -> 성공
- multi_code : 이전에 작업한 retry_code 중 발행자와 구독자가 합쳐져있던 클라이언트 코드 1개를 발행자와 구독자 코드 총 2개로 분해 및 명령행 인자 구현(command-line arguments)을 적용해 브로커 및 클라이언트 구동/연동 시도 -> 성공
