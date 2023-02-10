# kafka-transactional-producer

Demo a transactional producer with the confluent-kafka-go API.

The producer will atomically produce a 2 messages to a local kafka cluster.

First message is a dummy message.

Second message contains information about the first one.

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
