package amqp

import (
	"context"
	"errors"
	"strings"

	"github.com/XSAM/go-hybrid/log"
	"github.com/streadway/amqp"
)

var (
	// ErrClosed ...
	ErrClosed = amqp.ErrClosed
	// ErrDeliveryChannelClosed ...
	ErrDeliveryChannelClosed = errors.New("delivery channel was closed")
)

// Config ...
type Config amqp.Config

// Connection amqp connection abstraction.
type Connection struct {
	brokerURL string
	*amqp.Connection
}

// DialConfig ...
func DialConfig(url string, config Config) (*Connection, error) {
	c, err := amqp.DialConfig(url, (amqp.Config)(config))
	if err != nil {
		return nil, err
	}
	server := url
	if parts := strings.SplitN(url, "@", 2); len(parts) == 2 {
		server = parts[1]
	}
	return &Connection{Connection: c, brokerURL: server}, nil
}

// Channel get channel from a specific amqp connection.
func (c *Connection) Channel() (*Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, err
	}
	return &Channel{Channel: ch, c: c}, nil
}

type Publishing amqp.Publishing

type Error amqp.Error

// Channel amqp channel abstraction.
type Channel struct {
	*amqp.Channel
	c *Connection
}

// NotifyClose notify error to listener while channel is closing.
func (ch *Channel) NotifyClose(c chan *Error) chan *Error {
	errC := make(chan *amqp.Error)
	ch.Channel.NotifyClose(errC)
	go func() {
		for err := range errC {
			c <- (*Error)(err)
		}
	}()
	return c
}

// Publish publish a message.
func (ch *Channel) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg Publishing) (err error) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}
	return ch.Channel.Publish(exchange, key, mandatory, immediate, (amqp.Publishing)(msg))
}

type Delivery struct {
	*amqp.Delivery
}

func (d *Delivery) Ack(ctx context.Context, multiple bool) (err error) {
	return d.Delivery.Ack(multiple)
}

func (d *Delivery) Reject(ctx context.Context, requeue bool) (err error) {
	return d.Delivery.Reject(requeue)
}

// DeliveryArgs args which consuming from a queue.
type DeliveryArgs struct {
	// Queue name
	Queue string
	// ConsumerTag consumer tag, if empty name specified, rabbitmq will generate a random consumer tag.
	ConsumerTag string
	// AutoAck acknowledge message automatically.
	AutoAck bool
	// Exclusive indicates the queue is exclusive.
	Exclusive bool
	Nowait    bool
	// Args arguments.
	Args amqp.Table
}

// Delivery get delivery chan from underlying amqp channel.
func (ch *Channel) Delivery(args *DeliveryArgs) (<-chan amqp.Delivery, error) {
	return ch.Channel.Consume(args.Queue, args.ConsumerTag, args.AutoAck, args.Exclusive, false, args.Nowait, args.Args)
}

type Handler func(ctx context.Context, ch *Channel, d *Delivery) error

// Consume consume message from amqp broker in block mode.
func (ch *Channel) Consume(ctx context.Context, queue string, handler Handler, dc <-chan amqp.Delivery) error {
	log.BgLogger().Infof("amqp: start the consumer of %s queue", queue)
	for {
		select {
		case <-ctx.Done():
			log.BgLogger().Infof("amqp: stop the consumer of %s queue", queue)
			return nil
		case d, ok := <-dc:
			if !ok {
				log.BgLogger().Warnf("amqp: the deliver channel of %s queue closed", queue)
				return ErrDeliveryChannelClosed
			}
			go func() {
				err := ch.consume(queue, handler, &Delivery{&d})
				if err != nil {
					log.BgLogger().Warnf("amqp: execute handler with queue %s failed, reason: %v", queue, err.Error())
				}
			}()
		}
	}
}

func (ch *Channel) consume(queue string, handler Handler, d *Delivery) (err error) {
	return handler(context.Background(), ch, d)
}
