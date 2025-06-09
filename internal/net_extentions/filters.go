package net_extentions

func IsHttps(p TargetPeer) bool {
	return p.GetTargetPort() == 443
}
