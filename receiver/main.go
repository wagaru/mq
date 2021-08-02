package main

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs-topic",
		"topic",
		true, //durable
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false,
		true, //exclusive
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [binding_key]", os.Args[0])
		os.Exit(0)
	}

	for _, s := range os.Args[1:] {
		err = ch.QueueBind(
			q.Name,       // queue name
			s,            // routing key
			"logs-topic", // exchange
			false,
			nil,
		)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true, // auto-ack
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("[x] %s", d.Body)
		}
	}()

	log.Print(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s %s", err, msg)
	}
}
