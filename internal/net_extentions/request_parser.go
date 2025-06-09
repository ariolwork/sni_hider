package net_extentions

import (
	"strconv"
	"strings"
)

type TargetPeer interface {
	GetTargetUrl() string
	GetTargetPort() int
	GetTargetName() string
}

type peer struct {
	url  string
	port int
	name string
}

func (p *peer) GetTargetUrl() string {
	return p.url
}

func (p *peer) GetTargetPort() int {
	return p.port
}

func (p *peer) GetTargetName() string {
	return p.name
}

func ExtractTargetPeer(r []byte) (TargetPeer, error) {
	p := &peer{}
	header := string(r[:500])
	p.url = strings.Split(header, " ")[1]
	addrAndPort := strings.Split(p.url, ":")[:2]
	p.name = addrAndPort[0]
	port, err := strconv.Atoi(addrAndPort[1])
	p.port = port
	if err != nil {
		return nil, err
	}
	return p, nil
}
