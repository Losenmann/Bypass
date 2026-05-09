package firewall

import (
	"net"

	"github.com/lrh3321/ipset-go"
)

type IPSetManager struct {
	setname string
	typename string
}

func NewIPSetManager() *IPSetManager {
	return &IPSetManager{
		setname: "bypass",
		typename: ipset.TypeHashIP,
	}
}

func (m *IPSetManager) Setupe() {
	err := ipset.Create(m.setname, m.typename, ipset.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func (m *IPSetManager) Destroy() {
	err := ipset.Destroy(m.setname)
	if err != nil {
		panic(err)
	}
}

func (m *IPSetManager) Add() {
	err := ipset.Add(m.setname, &ipset.Entry{IP: net.IPv4(10, 0, 0, 1).To4()})
	if err != nil {
		panic(err)
	}
}

func (m *IPSetManager) Del() {
	err := ipset.Del(m.setname, &ipset.Entry{IP: net.IPv4(10, 0, 0, 1).To4()})
	if err != nil {
		panic(err)
	}
}


func RunIPSet() {
	ipsetManager := NewIPSetManager()
	ipsetManager.Setupe()
}
