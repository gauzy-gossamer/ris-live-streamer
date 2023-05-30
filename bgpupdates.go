package main

import (
    "time"
    "strings"
    "strconv"
    "log"
)

func convertTime(s string) float64 {
    t, err := time.Parse("2006-01-02 15:04:05", s)

    if err != nil {
        return 0
    }
    return float64(t.Unix())
}

func convertASPath(s string) []int {
    asns := strings.Split(s, " ")
    asns_int := []int{}
    for _, asn := range asns {
        asn_int, err := strconv.Atoi(asn)
        if err != nil {
            log.Println("asn conversion error", asn)
            continue
        }
        asns_int = append(asns_int, asn_int)
    }
    return asns_int
}

func convertCommunity(s string) [][]int {
    result := [][]int{}

    pairs := strings.Split(s, " ")
    for _, pair := range pairs {
        asns := strings.Split(pair, ":")
        if len(asns) != 2 {
            continue
        }
        from, err := strconv.Atoi(asns[0])
        if err != nil {
            continue
        }
        to, err := strconv.Atoi(asns[1])
        if err != nil {
            continue
        }

        result = append(result, []int{from, to})
    }

    return result
}

/* read messages data from channel provided by kafka consumer and send them to relevant subscribers */
func monitorBGPUpdates(ch <-chan KafkaMessage, subscriptions *Subscriptions) {
    bgpupdate := BGPUpdate{}
    prev_msg := KafkaMessage{}

    send_data := func() {
        conns := subscriptions.Filter(bgpupdate)
        if len(conns) == 0 {
            return
        }
        bgpmsg := RisData{Type:RIS_MESSAGE, Data:bgpupdate}
        for conn := range conns {
            if err := conn.SocketWriteJSON(bgpmsg); err != nil {
                log.Println(err)
            }
        }
    }

    for {
        msg := <-ch

        /* aggregate messages before sending them */
        if msg.Peer == prev_msg.Peer && msg.NextHop == prev_msg.NextHop &&
           msg.Type == prev_msg.Type && msg.ASPath == prev_msg.ASPath {
            if msg.Type == "withdraw" {
                bgpupdate.Withdrawals = append(bgpupdate.Withdrawals, msg.Prefix)
            } else {
                bgpupdate.Announcements = append(bgpupdate.Announcements, msg.Prefix)
            }
        } else {
            if bgpupdate.UpdateType != "" {
                send_data()
            }
            bgpupdate = BGPUpdate{
                Timestamp:convertTime(msg.Timestamp),
                UpdateType:strings.ToUpper(msg.Type),
                Peer:msg.Peer,
                Origin:msg.Origin,
                Community:convertCommunity(msg.Comms),
                Path:convertASPath(msg.ASPath),
            }
            if msg.Type == "withdraw" {
                bgpupdate.Announcements = []string{}
                bgpupdate.Withdrawals = []string{msg.Prefix}
            } else {
                bgpupdate.Announcements = []string{msg.Prefix}
                bgpupdate.Withdrawals = []string{}
            }
        }

        prev_msg = msg
    }
}
