package main

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}
const (
	SmokeDetected    = "1"
	NoSmokeDetected  = "0"
)

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		sensorStatus := string(msg.Payload)
		if err != nil {
			msg.Nack()
			continue
		}

		if sensorStatus == SmokeDetected {
			err := alarmClient.StartAlarm()
			if err != nil {
				msg.Nack()
				continue
			}
			msg.Ack()
		} else if sensorStatus == NoSmokeDetected {	
		
			err := alarmClient.StopAlarm()
			if err != nil {
				msg.Nack()
				continue
			}	
			msg.Ack()
		}


	}
}
