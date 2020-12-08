// +build integration

package amqp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DeBankDeFi/golib/util"

	"github.com/stretchr/testify/require"
)

var (
	// rabbitmq docker-compose: https://github.com/micahhausler/rabbitmq-compose
	brokerURL      = "amqp://rabbitmq:rabbitmq@127.0.0.1:5672/test"
	testExchange   = "test-exchange"
	testRoutingKey = "test.rk"
	testQueue      = "test-qq"
)

func TestDialConfig(t *testing.T) {
	conn, err := DialConfig(brokerURL, Config{
		Heartbeat: 10 * time.Second,
	})
	require.NoError(t, err)
	defer conn.Close()
}

func TestConnection_Channel(t *testing.T) {
	conn, err := DialConfig(brokerURL, Config{
		Heartbeat: 10 * time.Second,
	})
	require.NoError(t, err)
	defer conn.Close()
	channel, err := conn.Channel()
	require.NoError(t, err)
	defer channel.Close()
}

func TestChannel_Publish(t *testing.T) {
	conn, err := DialConfig(brokerURL, Config{
		Heartbeat: 10 * time.Second,
	})
	require.NoError(t, err)
	defer conn.Close()
	channel, err := conn.Channel()
	require.NoError(t, err)
	defer channel.Close()
	for i := 0; i < 100; i++ {
		bodyString := util.RandomName(64)
		channel.Publish(context.Background(), testExchange, testRoutingKey, false, false, Publishing{
			ContentType: "application/data",
			Body:        []byte(bodyString),
		})
	}
}

func TestChannel_Consume(t *testing.T) {
	conn, err := DialConfig(brokerURL, Config{
		Heartbeat: 10 * time.Second,
	})
	require.NoError(t, err)
	defer conn.Close()
	channel, err := conn.Channel()
	require.NoError(t, err)
	defer channel.Close()

	args := &DeliveryArgs{
		Queue:       testQueue,
		ConsumerTag: "tester-" + util.RandomName(16),
		AutoAck:     false,
		Exclusive:   false,
		Nowait:      false,
		Args:        nil,
	}
	dc, err := channel.Delivery(args)
	require.NoError(t, err)

	handler := func(ctx context.Context, ch *Channel, d *Delivery) error {
		defer d.Ack(ctx, false)
		fmt.Println("receive message from amqp: ", string(d.Body))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	go channel.Consume(ctx, testQueue, handler, dc)
	<-ctx.Done()
}
