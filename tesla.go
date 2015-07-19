package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/InfiniteMoments/tesla/controller"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type AMQP_config struct {
	uri, exchange, exchangeType, queue, bindingKey, consumerTag string
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

var stopChannel = make(chan string, 10)

func initConfig() {
	viper.SetConfigName("moments_config")
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

func initLog() {
	f, err := os.OpenFile("$GOPATH/src/github.com/tesla/log/mq_client", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
}

func main() {
	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	// initLog()
	initConfig()
	mq_config := initAMQP()

	c, err := NewConsumer(mq_config.uri, mq_config.exchange, mq_config.exchangeType, mq_config.queue, mq_config.bindingKey, mq_config.consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Wait for receiving a signal. This will clean exit on SIGINT, SIGKILL and SIGTERM and ensures the defers run.
	<-sigc

	log.Printf("Gracefully shutting down...")

	if err := c.Shutdown(); err != nil {
		log.Fatalf("error during shutdown: %s", err)
	}
}

// Connect queue to the exchange to receive string to be parsed
func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		log.Printf("RMQERR Dial: %s", err)
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		log.Printf("RMQERR Channel: %s", err)
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Printf("RMQERR Exchange Declare: %s", err)
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("RMQERR Queue Declare: %s", err)
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		log.Printf("RMQERR Queue Bind: %s", err)
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		log.Printf("RMQERR Queue Consume: %s", err)
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}
	go handle(deliveries, c.done)

	return c, nil
}

// Handle rabbitmq shutdown
func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		log.Printf("ERR Consumer cancel failed: %s", err)
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("ERR AMQP connection close error: %s", err)
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	log.Printf("AMQP shutdown OK")
	defer fmt.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

// Handle all strings received and pass it through to be processed
func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		Recvd := string(d.Body)
		parsedArray := strings.Split(Recvd, ":")
		switch parsedArray[0] {
		case "start":
			go controller.StartStream(parsedArray[1])
		case "stop":
			go controller.StopStream(parsedArray[1])
		}
		d.Ack(true)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
