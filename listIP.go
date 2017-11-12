package listIP

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"time"
)

// IPList : Struct to work with
type IPList struct {
	ip    []net.IP
	alive []string
	fail  []string
}

func (n *IPList) fill(ip net.IP) {
	novaIP := make(net.IP, len(ip))
	copy(novaIP, ip)
	n.ip = append(n.ip, novaIP)
}

/*
  fillNetwork: fills all network adresses
*/
func (n *IPList) fillNetwork(ip net.IP, ipnet *net.IPNet) {
	notfirst := false

	prefix := &net.IPNet{IP: ip.Mask(ipnet.Mask), Mask: ipnet.Mask}
	broadcastAddr := lastAddr(prefix)

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {

		if !notfirst {
			// remove network address
			notfirst = true
			continue
		}
		// Skip broadcast address
		if ip.Equal(broadcastAddr) {
			continue
		}
		novaIP := make(net.IP, len(ip))
		copy(novaIP, ip)

		n.ip = append(n.ip, novaIP)

	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// testAliveHosts: Checks if a host is Alive. fills an IPList struct
func (n *IPList) testAliveHosts(port int, timeout string) {

	messages, errc := make(chan string), make(chan error)

	numHosts := len(n.ip)

	for i := 0; i < numHosts; i++ {
		go isAlive(n.ip[i], port, timeout, messages, errc)
	}

	var alives []string
	var faileds []string

	for i := 0; i < len(n.ip); i++ {
		select {
		case res := <-messages:
			alives = append(alives, res)
		case err := <-errc:
			faileds = append(faileds, err.Error())
			// log.Println(err.Error())
		}
	}

	n.alive = alives
	n.fail = faileds
}

// isAlive : test if a IP answers in a port
func isAlive(ip net.IP, port int, timeout string, messages chan string, errc chan error) {
	timeoutDuration, err := time.ParseDuration(timeout)

	connexio := ip.String() + ":" + strconv.Itoa(port)

	conn, err := net.DialTimeout("tcp", connexio, timeoutDuration)
	if err != nil {
		errc <- err
		return
	}
	messages <- ip.String()
	conn.Close()
}

/* Determine last address of a range */
func lastAddr(n *net.IPNet) net.IP {
	ip := make(net.IP, len(n.IP.To4()))
	binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())|^binary.BigEndian.Uint32(net.IP(n.Mask).To4()))
	return ip
}

/*
Check a list of CIDR networks for the specified port open
*/
func Check(rangs []string, port int, timeout string) []string {
	var ips IPList

	for rang := range rangs {
		ip, ipnet, err := net.ParseCIDR(rangs[rang])
		if err == nil {
			ips.fillNetwork(ip, ipnet)
		} else {
			ip := net.ParseIP(rangs[rang])
			if ip != nil {
				ips.fill(ip)
			} else {
				log.Fatal("Address not in IP nor CIDR format:", rangs[rang])
			}
		}

	}

	ips.testAliveHosts(port, timeout)

	return ips.alive
}

/*
CheckFullStatus returns a IPList struct with alives and failed IPs in a range
*/
func CheckFullStatus(rangs []string, port int, timeout string) IPList {
	var ips IPList

	for rang := range rangs {
		ip, ipnet, err := net.ParseCIDR(rangs[rang])
		if err == nil {
			ips.fillNetwork(ip, ipnet)
		} else {
			ip := net.ParseIP(rangs[rang])
			if ip != nil {
				ips.fill(ip)
			} else {
				log.Fatal("Address not in IP nor CIDR format:", rangs[rang])
			}
		}

	}

	ips.testAliveHosts(port, timeout)

	return ips
}
