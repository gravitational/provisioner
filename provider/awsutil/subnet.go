package awsutil

import (
	"net"

	"github.com/gravitational/trace"
)

// SelectVPCSubnet returns a /24 subnet that does not overlap with the provided subnet blocks
// from the provided VPC block
func SelectVPCSubnet(vpcBlock string, subnetBlocks []string) (string, error) {
	_, vpcNet, err := net.ParseCIDR(vpcBlock)
	if err != nil {
		return "", trace.Wrap(err)
	}

	subnetNets, err := parseCIDRs(subnetBlocks)
	if err != nil {
		return "", trace.Wrap(err)
	}

	nextNet := net.IPNet{IP: vpcNet.IP, Mask: mask24}
	for vpcNet.Contains(nextNet.IP) {
		if intersects(nextNet, subnetNets) {
			nextNet = next24Net(nextNet.IP)
			continue
		}
		return nextNet.String(), nil
	}

	// we went out of bounds of VPC block, apparently it does not have a free /24 subnet
	return "", trace.NotFound("no /24 subnet found in %v", vpcBlock)
}

// parseCIDRs returns a list of IP networks parsed from the provided list
func parseCIDRs(blocks []string) ([]net.IPNet, error) {
	ipNets := make([]net.IPNet, 0, len(blocks))
	for _, block := range blocks {
		_, ipNet, err := net.ParseCIDR(block)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		ipNets = append(ipNets, *ipNet)
	}
	return ipNets, nil
}

// intersects returns true if the provided network "ipNet" intersects with any of
// the provided networks "ipNets"
func intersects(ipNet net.IPNet, ipNets []net.IPNet) bool {
	for _, n := range ipNets {
		if n.Contains(ipNet.IP) || ipNet.Contains(n.IP) {
			return true
		}
	}
	return false
}

// next24Net returns the next /24 subnet relative to the subnet the provided IP
// belongs to, e.g. if the provided IP is 10.100.1.1, the returned subnet will
// be 10.100.2.0/24
func next24Net(ip net.IP) net.IPNet {
	new := make(net.IP, len(ip))
	copy(new, ip)
	// increment starting from the 2nd octet from the right
	for i := len(new) - 2; i >= 0; i-- {
		new[i]++ // if overflown, increment the next octet to the left
		if new[i] > 0 {
			break
		}
	}
	return net.IPNet{IP: new, Mask: mask24}
}

var (
	// mask16 is the decimal netmask for a /16 subnet
	mask16 = net.IPv4Mask(255, 255, 0, 0)

	// mask24 is the decimal netmask for a /24 subnet
	mask24 = net.IPv4Mask(255, 255, 255, 0)

	// privateNets is the blocks of the IP address space reserved
	// for private internets (RFC1918):
	//   10.0.0.0 - 10.255.255.255  (10/8 prefix)
	//   172.16.0.0 - 172.31.255.255  (172.16/12 prefix)
	//   192.168.0.0 - 192.168.255.255 (192.168/16 prefix)
	privateNets = []net.IPNet{
		{
			IP:   net.IPv4(10, 0, 0, 0),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
		{
			IP:   net.IPv4(172, 16, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
		{
			IP:   net.IPv4(192, 168, 0, 0),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		},
	}
)
