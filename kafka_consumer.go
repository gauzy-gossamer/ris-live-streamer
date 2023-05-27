package main

import (
    "context"
    "log"
    "time"
    "encoding/json"

    "github.com/segmentio/kafka-go"
)

/* read from kafka and write messages to ch that's listened by monitorBGPUpdates */
func processKafkaMessages(ch chan<- KafkaMessage, config KafkaConfig) {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:   config.Brokers,
        GroupID:   config.GroupID,
        Topic:     config.Topic,
        Partition: 0,
        MinBytes:  10e3, //10KB
        MaxBytes:  10e6, // 10MB
        MaxWait:   1*time.Second,
        CommitInterval: 500 * time.Millisecond,
    })

    for {
        m, err := r.ReadMessage(context.Background())
        if err != nil {
            break
        }

        msg := KafkaMessage{}
        err = json.Unmarshal(m.Value, &msg)
        if err != nil {
            log.Print(err)
            continue
        }
        ch <- msg
    }

    if err := r.Close(); err != nil {
        log.Fatal("failed to close reader:", err)
    }
}
