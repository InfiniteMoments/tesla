package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"log"
)

type connection struct {
	mq *amqp.Connection
}

var conn connection

type AMQP_config struct {
	uri, exchange, exchangeType, queue, bindingKey, consumerTag string
}

func initConfig() {
	viper.SetConfigName("moments_config")
	viper.AddConfigPath("$GOPATH/src/github.com/InfiniteMoments/tesla")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

}

func initAMQP() AMQP_config {
	return AMQP_config{
		"amqp://" + viper.GetString("RMQP_USERNAME") + ":" + viper.GetString("RMQP_PASSWORD") + "@localhost:5672/",
		viper.GetString("RMQP_EXCHANGE"),
		viper.GetString("RMQP_EXCHANGETYPE"),
		viper.GetString("RMQP_QUEUE"),
		viper.GetString("RMQP_BINDINGKEY"),
		viper.GetString("RMQP_CONSUMERTAG"),
	}
}

var done = make(chan string, 1)

func main() {
	initConfig()
	mq_config := initAMQP()

	go publish(mq_config.uri, mq_config.exchange, mq_config.exchangeType, mq_config.bindingKey, true)

	for {
		select {
		case res := <-done:
			fmt.Println("Published", res)
			return
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// Connect to Potgres, get all MAC IDs in the homemeter table and publish it to the exchange
func publish(amqpURI, exchange, exchangeType, routingKey string, reliable bool) error {
	var err error

	var searchQuery string
	fmt.Println("Stream twitter for : ")
	fmt.Scan(&searchQuery)

	var queryType string
	fmt.Println("start/stop : ")
	fmt.Scan(&queryType)

	log.Printf("dialing %q", amqpURI)
	conn.mq, err = amqp.Dial(amqpURI)
	if err != nil {
		log.Printf("RMQERR Dial: %s", err)
		return fmt.Errorf("RMQERR Dial: %s", err)
	}
	defer conn.mq.Close()

	log.Printf("got Connection, getting Channel")
	channel, err := conn.mq.Channel()
	if err != nil {
		log.Printf("RMQERR Channel: %s", err)
		return fmt.Errorf("RMQERR Channel: %s", err)
	}

	log.Printf("got Channel, declaring %q Exchange (%q)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Printf("RMQERR Exchange Declare: %s", err)
		return fmt.Errorf("RMQERR Exchange Declare: %s", err)
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if reliable {
		log.Printf("enabling publishing confirms.")
		if err := channel.Confirm(false); err != nil {
			log.Printf("RMQERR Channel could not be put into confirm mode: %s", err)
			fmt.Errorf("RMQERR Channel could not be put into confirm mode: %s", err)
		}
	}
	ack, nack := channel.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	var body string

	body = queryType + ":" + searchQuery

	defer confirmOne(ack, nack)

	log.Printf("publishing %dB body (%q)", len(body), body)
	if err = channel.Publish(
		exchange,   // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		log.Printf("RMQERR Exchange Publish: %s", err)
		return fmt.Errorf("RMQERR Exchange Publish: %s", err)
	}

	return nil
}

// Confirmation for each MAC ID sent
func confirmOne(ack, nack chan uint64) {
	log.Printf("waiting for confirmation of one publishing\n")

	select {
	case tag := <-ack:
		{
			log.Printf("confirmed delivery with delivery tag: %d", tag)
			done <- "OK"
		}
	case tag := <-nack:
		log.Printf("failed delivery of delivery tag: %d", tag)
	}
}
