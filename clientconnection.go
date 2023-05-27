package main

import (
    "sync"
    "time"
    "log"
    "encoding/json"

    "github.com/gorilla/websocket"
)

const (
    closeGracePeriod = 10 * time.Second
    readTimeout = 120* time.Second
)

type ClientConnection struct {
    conn *websocket.Conn
    mu sync.Mutex
    subscriptions *Subscriptions

    /* indicated that the socket is closed */
    closed chan bool

    last_received time.Time
    last_sent time.Time
}

func NewClientConnection(conn *websocket.Conn, subs *Subscriptions) ClientConnection {
    cconn := ClientConnection{conn:conn, subscriptions:subs}
    cconn.closed = make(chan bool)
    cconn.last_sent = time.Now()
    return cconn
}

func (c *ClientConnection) SocketWriteJSON(message RisData) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.last_sent = time.Now()
    message_resp, err := json.Marshal(message)
    if err != nil {
        return err
    }
    return c.conn.WriteMessage(websocket.TextMessage, message_resp)
}

func (c *ClientConnection) SocketWrite(message []byte) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.last_sent = time.Now()
    return c.conn.WriteMessage(websocket.TextMessage, message)
}

func getRisError(message string) RisData {
    response := RisData{}
    response.Type = "ris_error"
    response.Data = ErrMessage{Message:message}

    return response
}

func (c *ClientConnection) ProcessCommand(cmdType string, data interface{}) RisData {
    response := RisData{}
    switch cmdType {
    case "ping":
        response.Type = "pong"
    case "ris_subscribe":
        if sub, ok := data.(map[string]interface{}) ;  ok {
            response.Type = "ris_subscribe_ok"
            c.subscriptions.Add(c.subscriptions.ParseSub(sub), c)

        } else {
            response = getRisError("unknown command")
        }
    case "ris_unsubscribe":
        if sub, ok := data.(map[string]interface{}) ;  ok {
            response.Type = "ris_unsubscribe_ok"
            deleted := c.subscriptions.Delete(c.subscriptions.ParseSub(sub), c)
            if !deleted {
                response = getRisError("subscription not found")
            }

        } else {
            response = getRisError("unknown command")
        }
    default:
        response = getRisError("unknown command")
    }

    return response
}

func (c *ClientConnection) ConnLoop() {
    c.conn.SetCloseHandler(func (code int, text string) error {
        message := websocket.FormatCloseMessage(code, "")
        c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
        c.subscriptions.DeleteAll(c)
        log.Println(code)
        close(c.closed)
        return nil
    })

    go func() {
        for {
            time.Sleep(time.Second*5)
            if c.isClosed() {
                break
            }

            if time.Since(c.last_sent) > readTimeout && !c.subscriptions.HasSubscriptions(c) {
                log.Println("close idle connection", time.Since(c.last_sent))

                err := c.SocketWriteJSON(getRisError("Closing due to idle connection (no subscriptions)"))
                if err != nil {
                    log.Println("write :", err)
                }
                c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
                time.Sleep(closeGracePeriod)
                c.conn.Close()
                break
            }
        }
    }()

    for {
        data := RisData{}
        err := c.conn.ReadJSON(&data)
        if err != nil {
            if c.isClosed() {
                break
            }
            err := c.SocketWriteJSON(getRisError("Could not parse JSON"))
            if err != nil {
                if err == websocket.ErrCloseSent {
                    break
                }
                log.Println("write :", err)
            }
            continue
        }
        c.last_received = time.Now()
        response := c.ProcessCommand(data.Type, data.Data)
        err = c.SocketWriteJSON(response)
        if err != nil {
            log.Println("write:", err)
            break
        }
    }
}

func (c *ClientConnection) isClosed() bool {
    select{
    case _, ok := <-c.closed:
        return !ok
    default:
        return false
    }
}

func (c *ClientConnection) Close() {
    c.conn.Close()
}
