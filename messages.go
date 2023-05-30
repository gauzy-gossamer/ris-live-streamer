package main

var MESSAGE_PREFIX string = "ris"

var RIS_ERROR string = MESSAGE_PREFIX + "_error"
var RIS_SUBSCRIBE string = MESSAGE_PREFIX + "_subscribe"
var RIS_UNSUBSCRIBE string = MESSAGE_PREFIX + "_unsubscribe"
var RIS_SUBSCRIBE_OK string = MESSAGE_PREFIX + "_subscribe_ok"
var RIS_UNSUBSCRIBE_OK string = MESSAGE_PREFIX + "_unsubscribe_ok"
var RIS_MESSAGE string = MESSAGE_PREFIX + "_message"

/*
message received from Kafka
*/
type KafkaMessage struct {
    Timestamp string `json:"timestamp"`
    Peer string      `json:"peer_ip_src"`
    Prefix string    `json:"ip_prefix"`
    NextHop string   `json:"bgp_nexthop"`
    ASPath string    `json:"as_path"`
    Comms string     `json:"comms"`
    Origin string    `json:"origin"`
    Type string      `json:"log_type"`
}

/*
bgp update output message format 
*/
type BGPUpdate struct {
    Timestamp float64  `json:"timestamp"`
    Peer string        `json:"peer"`
    PeerASN string     `json:"peer_asn,omitempty"`
    Id string          `json:"id,omitempty"`
    Host string        `json:"host,omitempty"`
    UpdateType string  `json:"type"`
    Origin string      `json:"origin"`
    Path []int         `json:"path"`
    Community [][]int  `json:"community,omitempty"`
    Announcements []string `json:"announcements"`
    Withdrawals []string   `json:"withdrawals"`
}

type ErrMessage struct {
    Message string `json:"message"`
}

/* websocket message format 
data could either be ErrMessage or BGPUpdate */
type RisData struct {
    Type string      `json:"type"`
    Data any         `json:"data"`
}
