package main

import (
    "os"
    "strings"
    "bufio"
    "encoding/json"
)

type KafkaTopic struct {
    GroupID  string   `json:"group_id"`
    Topic    string   `json:"topic"`
}

type KafkaConfig struct {
    Brokers  []string `json:"brokers"`
    Topics []KafkaTopic `json:"topics"`
}

type Config struct {
    Host string       `json:"host"`
    WSPath string     `json:"ws_path"`
    Kafka KafkaConfig `json:"kafka"`
    Listen string     `json:"listen"`
    Prefix string     `json:"cmd_prefix,omitempty"`
}

func ReadJSON(filename string, v interface{}) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    return decoder.Decode(v)
}

func ReadConfig(filename string) (Config, error) {
    config := Config{}
    err := ReadJSON(filename, &config)
    if config.Prefix != "" {
        setMessages(config.Prefix)
    }
    return config, err
}

func readTemplate() (string, error) {
    file, err := os.Open("websocket.html")
    if err != nil {
        return "", err
    }

    templateContent := strings.Builder{}
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        templateContent.WriteString(scanner.Text())
        templateContent.WriteByte('\n')
    }

    return templateContent.String(), scanner.Err()
}

func setMessages(prefix string) {
    MESSAGE_PREFIX = prefix

    RIS_ERROR = MESSAGE_PREFIX + "_error"
    RIS_SUBSCRIBE = MESSAGE_PREFIX + "_subscribe"
    RIS_UNSUBSCRIBE = MESSAGE_PREFIX + "_unsubscribe"
    RIS_SUBSCRIBE_OK = MESSAGE_PREFIX + "_subscribe_ok"
    RIS_UNSUBSCRIBE_OK = MESSAGE_PREFIX + "_unsubscribe_ok"
    RIS_MESSAGE = MESSAGE_PREFIX + "_message"
}
