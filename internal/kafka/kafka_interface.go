//go:generate mockgen -source ./kafka_interface.go -destination=./mocks/kafka_interface.go -package=mock_kafka_interface

package kafka

type KafkaInterface interface {
	SendEvent(topic string, event Event) error
}
