# kafka-transactional-producer

Demo a transactional producer with the confluent-kafka-go API.

The producer will atomically send 2 messages to a local kafka cluster.

The first message is a dummy message.

The second message contains information about the first one.

# Getting Started

Start the kafka cluster
```
make kafka
```

Create the topics
```
make topics
```

Run the producer
```
make run
```
