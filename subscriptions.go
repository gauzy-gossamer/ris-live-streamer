package main

import (
    "sync"
    "errors"
    "strings"
    "reflect"
)

type ClientConnectionI interface {
    SocketWriteJSON(message RisData) error
}

func setField(v interface{}, name string, value any) error {
    fv := reflect.ValueOf(v).Elem().FieldByName(name)

    switch value := value.(type) {
        case string:
            fv.SetString(value)
        case uint:
            fv.SetUint(uint64(value))
        case int:
            fv.SetInt(int64(value))
        case bool:
            fv.SetBool(value)
        default:
            return errors.New("unknown type")
    }
    return nil
}

type Sub struct {
    Host string
    Path string
    Peer string
    Prefix string
    Require string
    UpdateType string
}

type Subscriptions struct {
    mu sync.RWMutex

    /* host, peer, prefix, path */
    specific_subs map[string]map[string][]*Sub

    /* type, require */
    broad_subs []*Sub

    subs map[ClientConnectionI]map[Sub]struct{}
    subs_map map[Sub]map[ClientConnectionI]struct{}
}

func NewSubscriptions() *Subscriptions {
    subs := &Subscriptions{}
    subs.subs = make(map[ClientConnectionI]map[Sub]struct{})
    subs.subs_map = make(map[Sub]map[ClientConnectionI]struct{})
    subs.specific_subs = make(map[string]map[string][]*Sub)
    subs.specific_subs["host"] = make(map[string][]*Sub)
    subs.specific_subs["peer"] = make(map[string][]*Sub)
    subs.specific_subs["prefix"] = make(map[string][]*Sub)

    return subs
}

func (s *Subscriptions) ParseSub(vals map[string]interface{}) Sub {
    sub := Sub{}
    for k, v := range vals {
        if k == "type" {
            setField(&sub, "UpdateType", v)
        } else {
            setField(&sub, strings.Title(k), v)
        }
    }
    return sub
}

func (s *Subscriptions) Add(sub Sub, conn ClientConnectionI) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, found := s.subs[conn]; !found {
        s.subs[conn] = make(map[Sub]struct{})
    }

    if _, found := s.subs[conn][sub]; !found {
        s.subs[conn][sub] = struct{}{}
    }

    if _, found := s.subs_map[sub]; !found {
        s.subs_map[sub] = make(map[ClientConnectionI]struct{})
    }
    if _, found := s.subs_map[sub][conn]; !found {
        s.subs_map[sub][conn] = struct{}{}
    }
}

func (s *Subscriptions) Delete(sub Sub, conn ClientConnectionI) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, found := s.subs[conn]; !found {
        return false
    }

    if _, found := s.subs[conn][sub]; found {
        delete(s.subs[conn], sub)
        delete(s.subs_map[sub], conn)
    }
    if len(s.subs[conn]) == 0 {
        delete(s.subs, conn)
    }
    if len(s.subs_map[sub]) == 0 {
        delete(s.subs_map, sub)
    }
    return true
}

func (s *Subscriptions) DeleteAll(conn ClientConnectionI) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, found := s.subs[conn]; !found {
        return
    }

    for sub, _ := range s.subs[conn] {
        delete(s.subs_map[sub], conn)
        if len(s.subs_map[sub]) == 0 {
            delete(s.subs_map, sub)
        }
    }

    if len(s.subs[conn]) != 0 {
        delete(s.subs, conn)
    }
}

func (s *Subscriptions) HasSubscriptions(conn ClientConnectionI) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()

    _, found := s.subs[conn]
    return found
}

func SubMatch(sub *Sub, bgpupdate *BGPUpdate) bool {
    if sub.Host != "" && sub.Host != bgpupdate.Host {
        return false
    }
    if sub.Peer != "" && sub.Peer != bgpupdate.Peer {
        return false
    }
    if sub.Prefix != "" {
//        return false
    }
    if sub.UpdateType != "" && sub.UpdateType != bgpupdate.UpdateType {
        return false
    }
    if sub.Require == "announcements" && len(bgpupdate.Announcements) == 0 {
        return false
    }
    if sub.Require == "withdrawals" && len(bgpupdate.Withdrawals) == 0 {
        return false
    }

    return true
}

func (s *Subscriptions) Filter(bgpupdate BGPUpdate) map[ClientConnectionI]struct{} {
    s.mu.RLock()
    defer s.mu.RUnlock()

    conns := make(map[ClientConnectionI]struct{})

    for sub, conns_ := range s.subs_map {
        if !SubMatch(&sub, &bgpupdate) {
            continue
        }
        for conn, _ := range conns_ {
            conns[conn] = struct{}{}
        }
    }

    return conns
}
