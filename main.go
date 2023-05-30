package main

import (
    "flag"
    "html/template"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

type Server struct {
    config Config
    subscriptions *Subscriptions
}

var server Server

var upgrader = websocket.Upgrader{}

var homeTemplate *template.Template

func WSHandler(w http.ResponseWriter, r *http.Request) {
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("upgrade:", err)
        return
    }
    cconn := NewClientConnection(c, server.subscriptions)
    defer cconn.Close()

    cconn.ConnLoop()
}

func home(w http.ResponseWriter, r *http.Request) {
    homeTemplate.Execute(w, "ws://"+server.config.Host+server.config.WSPath)
}

func main() {
    config_file := flag.String("config", "config.json", "filename with config")
    addr := flag.String("addr", "", "http service address")
    flag.Parse()

    log.SetFlags(0)

    config, err := ReadConfig(*config_file)
    if err != nil {
        log.Fatal("failed to read config:", err)
    }

    if *addr != "" {
        config.Listen = *addr
    }

    server = Server{config:config}

    templateContent, err := readTemplate()
    if err != nil {
        log.Print("read template:", err)
    }
    homeTemplate = template.Must(template.New("").Parse(templateContent))

    kafka_ch := make(chan KafkaMessage, 10)

    server.subscriptions = NewSubscriptions()

    for _, topic := range config.Kafka.Topics {
        go processKafkaMessages(kafka_ch, config.Kafka, topic)
    }
    go monitorBGPUpdates(kafka_ch, server.subscriptions)

    http.HandleFunc(config.WSPath, WSHandler)
    http.HandleFunc("/", home)

    log.Fatal(http.ListenAndServe(config.Listen, nil))
}
