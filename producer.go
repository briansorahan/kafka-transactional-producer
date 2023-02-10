package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const broker = "kafka1:9092"

var (
	consumerOffsetsTopic = "consumer_offsets_topic"
	fooTopic             = "foo_topic"
)

func main() {
	die(transactional())
}

type TransactionFunc func(context.Context, *kafka.Producer) error

func executeTransaction(ctx context.Context, producer *kafka.Producer, f TransactionFunc) error {
	if err := producer.BeginTransaction(); err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	if err := f(ctx, producer); err != nil {
		if arr := producer.AbortTransaction(ctx); arr != nil {
			return fmt.Errorf("aborting transaction: %w", arr)
		}
		return fmt.Errorf("error in producer transaction function: %w", err)
	}
	if err := producer.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func transactional() error {
	ctx := context.Background()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"transactional.id":  "foo",
	})
	if err != nil {
		return fmt.Errorf("initializing producer: %w", err)
	}
	if err := producer.InitTransactions(ctx); err != nil {
		return fmt.Errorf("InitTransactions: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)

	defer cancel()

	if err := executeTransaction(ctx, producer, doTransaction); err != nil {
		return fmt.Errorf("executeTransaction: %w", err)
	}
	return nil
}

func doTransaction(ctx context.Context, producer *kafka.Producer) error {
	deliveryReports := producer.Events()

	// Produce a message on "foo_topic", wait for the delivery report, then
	// produce a message to the consumer offsets topic to track the offset
	// we produced on foo.
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &fooTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(`k`),
		Value: []byte(`v`),
	}
	fmt.Println("producing message to foo topic")

	if err := producer.Produce(msg, nil); err != nil {
		return fmt.Errorf("producing message to foo topic: %w", err)
	}
	fmt.Println("waiting for delivery report from foo topic")

	var (
		err    error
		offset kafka.Offset
	)
	select {
	case recv := <-deliveryReports:
		offset, err = getProducedOffset(recv)
		if err != nil {
			return fmt.Errorf("getting produced offset: %w", err)
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	offsetsMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &consumerOffsetsTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(`k`),
		Value: []byte(fmt.Sprintf(`produced offset %d`, offset)),
	}
	fmt.Println("producing message on consumer offsets topic")

	if err := producer.Produce(offsetsMsg, nil); err != nil {
		return fmt.Errorf("producing message on consumer offsets topic: %w", err)
	}
	fmt.Println("waiting for delivery report from consumer offsets topic")

	select {
	case recv := <-deliveryReports:
		if err := handle(recv); err != nil {
			return fmt.Errorf("handling event: %w", err)
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func getProducedOffset(event kafka.Event) (kafka.Offset, error) {
	switch x := event.(type) {
	case *kafka.Message:
		if err := x.TopicPartition.Error; err != nil {
			return 0, fmt.Errorf("error detected in delivery report: %w", err)
		}
		var (
			offset    = x.TopicPartition.Offset
			partition = x.TopicPartition.Partition
			topicName = *x.TopicPartition.Topic
		)
		fmt.Printf("produced message to %s [%d][%d]\n", topicName, partition, offset)

		return offset, nil
	case kafka.Error:
		return 0, fmt.Errorf("kafka error: %w", x)
	default:
		return 0, fmt.Errorf("unrecognized event type: %T", x)
	}
	return 0, fmt.Errorf("unreachable")
}

func handle(event kafka.Event) error {
	switch x := event.(type) {
	case *kafka.Message:
		if err := x.TopicPartition.Error; err != nil {
			return fmt.Errorf("error detected in delivery report: %w", err)
		}
		var (
			offset    = x.TopicPartition.Offset
			partition = x.TopicPartition.Partition
			topicName = *x.TopicPartition.Topic
		)
		fmt.Printf("produced message to %s [%d][%d]\n", topicName, partition, offset)
	case kafka.Error:
		return fmt.Errorf("kafka error: %w", x)
	default:
		return fmt.Errorf("unrecognized event type: %T", x)
	}
	return nil
}

func die(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}
