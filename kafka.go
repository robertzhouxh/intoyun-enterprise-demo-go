package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"os/signal"
	"time"

	cluster "github.com/bsm/sarama-cluster"
	log "github.com/thinkboy/log4go"
)

const (
	OFFSETS_PROCESSING_TIMEOUT_SECONDS = 10 * time.Second
	OFFSETS_COMMIT_INTERVAL            = 10 * time.Second
)

func md5password(id string, secret string) (password string) {
	h := md5.New()
	io.WriteString(h, secret)
	smd5 := hex.EncodeToString(h.Sum(nil))
	h1 := md5.New()
	io.WriteString(h1, id+string(smd5[:]))
	password = hex.EncodeToString(h1.Sum(nil))
	return password
}

func InitKafka() error {
	log.Info("start topic:%s consumer", Conf.KafkaTopic)
	log.Info("consumer group name:%s", Conf.Group)

	// init consumer
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	// for production please uncomment blows
	config.Net.SASL.Enable = Conf.SaslEnable
	config.Net.SASL.User = Conf.SaslUser
	config.Net.SASL.Password = md5password(Conf.SaslUser, Conf.SaslPassword)

	brokers := Conf.KafkaAddrs
	topics := []string{Conf.KafkaTopic}
	consumer, err := cluster.NewConsumer(brokers, Conf.Group, topics, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume messages, watch errors and notifications
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				//fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
				push(msg.Value)
				consumer.MarkOffset(msg, "") // mark message as processed
			}
		case err, more := <-consumer.Errors():
			if more {
				log.Error("Error: %s\n", err.Error())
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				log.Info("Info: %+v\n", ntf)
			}
		case <-signals:
			return nil
		}
	}

	return nil
}
