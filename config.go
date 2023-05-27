package main

import (
    "os"
    "strings"
    "bufio"
    "encoding/json"
)

type KafkaConfig struct {
    Brokers  []string `json:"brokers"`
    GroupID  string   `json:"group_id"`
    Topic    string   `json:"topic"`
}

type Config struct {
    Host string       `json:"host"`
    WSPath string     `json:"ws_path"`
    Kafka KafkaConfig `json:"kafka"`
    Listen string     `json:"listen"`
}

func ReadJSON(filename string, v interface{}) error {
    file, err := os.Open(filename)
    defer file.Close()
    if err != nil {
        return err
    }
    decoder := json.NewDecoder(file)
    return decoder.Decode(v)
}

func ReadConfig(filename string) (Config, error) {
    config := Config{}
    err := ReadJSON(filename, &config)
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