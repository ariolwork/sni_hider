package net_extentions

import (
	"context"
	"strconv"
	"strings"

	"tcp_sni_splitter/internal/net_extentions/dns"
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
	return p.name + ":" + strconv.Itoa(p.port)
}

func (p *peer) GetTargetPort() int {
	return p.port
}

func (p *peer) GetTargetName() string {
	return p.name
}

func ExtractTargetPeer(ctx context.Context, r []byte, resolver dns.Resolver) (TargetPeer, error) {
	p := &peer{}
	header := string(r[:500])
	p.url = strings.Split(header, " ")[1]
	addrAndPort := strings.Split(p.url, ":")[:2]
	p.name = addrAndPort[0]
	if p.name == "www.youtube.com" || p.name == "www.instagram.com" {
		p.name = resolver.Resolve(ctx, p.name)
		//p.name = "172.217.18.206"
	}

	port, err := strconv.Atoi(addrAndPort[1])
	p.port = port
	if err != nil {
		return nil, err
	}
	return p, nil
}
