package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"time"
)

type Event struct {
	TimeStamp   time.Time
	Type        string
	RequestBody string
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokerAddress string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &KafkaProducer{
		producer: producer,
	}, nil
}

func (producer *KafkaProducer) Close() error {
	err := producer.producer.Close()
	if err != nil {
		return fmt.Errorf("failed to close producer")
	}
	return nil
}

func (producer *KafkaProducer) SendEvent(topic string, event Event) error {
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder(fmt.Sprintf("EventType: %s, EventRequestBody: %s, EventTime: %v",
			event.Type, event.RequestBody, event.TimeStamp)),
	}

	_, _, err := producer.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %v", err)
	}

	return nil
}
