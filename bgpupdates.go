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

/* read messages data from channel provided by kafka consumer and send them to relevant subscribers */
func monitorBGPUpdates(ch <-chan KafkaMessage, subscriptions *Subscriptions) {
    bgpupdate := BGPUpdate{}
    prev_msg := KafkaMessage{}

    send_data := func() {
        conns := subscriptions.Filter(bgpupdate)
        if len(conns) == 0 {
            return
        }
        bgpmsg := RisData{Type:"ris_message", Data:bgpupdate}
        for conn, _ := range conns {
            conn.SocketWriteJSON(bgpmsg)
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
