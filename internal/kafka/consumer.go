package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"sync"
)

type KafkaConsumer struct {
	Consumer sarama.Consumer
}

func NewKafkaConsumer(brokerAddress string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{brokerAddress}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consumer: %w", err)
	}

	return &KafkaConsumer{
		Consumer: consumer,
	}, nil
}

func (consumer *KafkaConsumer) Close() error {
	err := consumer.Consumer.Close()
	if err != nil {
		return fmt.Errorf("failed to close Consumer")
	}
	return nil
}

func (consumer *KafkaConsumer) Consume(ctx context.Context, topic string) error {
	partitions, err := consumer.Consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to create Consumer")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	for _, partition := range partitions {
		pc, err := consumer.Consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to create Consumer for partition %d: %v", partition, err)
		}
		defer pc.Close()
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for {
				select {
				case message := <-pc.Messages():
					fmt.Println(string(message.Value))
				case <-ctx.Done():
					return
				}
			}
		}(pc)
	}
	wg.Wait()
	return err
}
