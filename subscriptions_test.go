package main

import (
    "testing"
)

type ClientConnectionM struct {
}

func (c *ClientConnectionM) SocketWriteJSON(message RisData) error {
    return nil
}

var test_bgpupdates = []BGPUpdate{
    {1684511834.89, "194.68.123.157", "9002", "194.68.123.157-018834bac30a0000", "rrc07.ripe.net", "UPDATE", "", []int{}, []string{}, []string{}, []string{"216.87.243.0/24", "216.87.248.0/24", "216.87.247.0/24", "216.87.240.0/20", "193.105.156.0/24"}},
    {1684511834.89, "194.68.123.157", "9002", "194.68.123.157-018834bac30a0000", "rrc08.ripe.net", "UPDATE", "", []int{}, []string{}, []string{"216.87.243.0/24", "216.87.248.0/24", "216.87.247.0/24", "216.87.240.0/20", "193.105.156.0/24"}, []string{}},
    {1684511834.89, "194.68.123.156", "9002", "194.68.123.157-018834bac30a0000", "rrc08.ripe.net", "UPDATE", "", []int{}, []string{}, []string{"216.87.243.0/24", "216.87.248.0/24", "216.87.247.0/24", "216.87.240.0/20", "193.105.156.0/24"}, []string{}},
}

func TestSubAddDelete(t *testing.T) {
    subs := NewSubscriptions()
    cconn := &ClientConnectionM{}
    sub := Sub{Host:"hh"}

    if subs.HasSubscriptions(cconn) {
        t.Error("shouldn't have subscriptions")
    }

    deleted := subs.Delete(sub, cconn)
    if deleted {
        t.Error("shouldn't have subscriptions")
    }

    subs.Add(sub, cconn)
    if !subs.HasSubscriptions(cconn) {
        t.Error("should have subscriptions")
    }

    deleted = subs.Delete(sub, cconn)
    if !deleted {
        t.Error("should have subscriptions")
    }

    subs.Add(sub, cconn)
    subs.Add(Sub{Host:"hh2"}, cconn)

    deleted = subs.Delete(sub, cconn)
    if !deleted {
        t.Error("should have subscriptions")
    }

    if !subs.HasSubscriptions(cconn) {
        t.Error("should have subscriptions")
    }
}

func TestSubFilter(t *testing.T) {
    subs := NewSubscriptions()
    cconn := &ClientConnectionM{}
    cconn2 := &ClientConnectionM{}
    sub := Sub{Host:"rrc07.ripe.net"}

    subs.Add(sub, cconn)

    conns := subs.Filter(test_bgpupdates[0])

    if len(conns) == 0 {
        t.Error("should receive filter", conns, test_bgpupdates[0])
    }

    subs.Add(Sub{Host:"rrc08.ripe.net"}, cconn)
    subs.Add(Sub{Host:"rrc08.ripe.net"}, cconn2)

    conns = subs.Filter(test_bgpupdates[0])

    if len(conns) != 1 {
        t.Error("should receive one conn", conns, test_bgpupdates[0])
    }

    subs.DeleteAll(cconn)
    if subs.HasSubscriptions(cconn) {
        t.Error("shouldn't have subscriptions")
    }
    subs.DeleteAll(cconn2)
    if subs.HasSubscriptions(cconn2) {
        t.Error("shouldn't have subscriptions")
    }
}

func BenchmarkSubFilter(b *testing.B) {
    subs := NewSubscriptions()
    cconn := &ClientConnectionM{}
    sub := Sub{Host:"rrc07.ripe.net"}
    subs.Add(sub, cconn)
    for i := 0; i < 100; i++ {
        cconn2 := &ClientConnectionM{}
        sub = Sub{Host:"rrc08.ripe.net"}
        subs.Add(sub, cconn2)
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        subs.Filter(test_bgpupdates[0])
    }
}

