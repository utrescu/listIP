package listIP

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"
)

// Number of parallel connections
const paralel = 32

// IPList : Struct to work with
type IPList struct {
	ip    []net.IP
	alive []string
	fail  []string
}

// fillIP :	Adds a sigle IP to IPList
func (n *IPList) fillIP(ip net.IP) {
	if ip == nil {
		return
	}
	novaIP := make(net.IP, len(ip))
	copy(novaIP, ip)
	n.ip = append(n.ip, novaIP)
}

//  fillNetwork: fills network adresses in IPList
func (n *IPList) fillNetwork(ip net.IP, ipnet *net.IPNet) {

	// skip nil values
	if ip == nil || ipnet == nil {
		return
	}

	// It's a single IP ...
	if ipnet.Mask.String() == net.CIDRMask(32, 32).String() {
		n.fillIP(ip)
		return
	}

	prefix := &net.IPNet{IP: ip.Mask(ipnet.Mask), Mask: ipnet.Mask}
	broadcastAddr := lastAddr(prefix)

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {

		// Skip broadcast and network addresses
		if ip.Equal(broadcastAddr) || ip.Equal(ipnet.IP) {
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
func (n *IPList) checkAliveHosts(port int, parallelconnections int, timeout time.Duration) {

	numHosts := len(n.ip)
	messages, errc := make(chan string, numHosts), make(chan error, numHosts)

	for i := 0; i < numHosts; i++ {
		go alive(n.ip[i], port, timeout, messages, errc)
		if i%parallelconnections == 0 {
			time.Sleep(timeout)
		}
	}

	var alives []string
	var faileds []string

	for i := 0; i < numHosts; i++ {
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

var dial = net.DialTimeout
var alive = isAlive

// isAlive : test if a IP answers in a port
func isAlive(ip net.IP, port int, timeout time.Duration, messages chan string, errc chan error) {

	connexio := ip.String() + ":" + strconv.Itoa(port)

	conn, err := dial("tcp", connexio, timeout)
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

params:
  - rangs: List of IP addresses or ranges in CIDR format
  - port: Port to test
  - parallelconnections: Number of parallel connections
  - timeout: defines the network timeout

returns:
  - list of alive hosts
  - list of failed hosts
*/
func Check(rangs []string, port int, parallelconnections int, timeout string) ([]string, []string, error) {
	var ips IPList

	if len(rangs) == 0 {
		return nil, nil, errors.New("No IPs provided")
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, nil, errors.New("Incorrect timeout")
	}

	if port <= 0 {
		return nil, nil, errors.New("Incorrect port ")
	}

	for rang := range rangs {
		ip, ipnet, err := net.ParseCIDR(rangs[rang])
		if err == nil {
			ips.fillNetwork(ip, ipnet)
		} else {
			ip := net.ParseIP(rangs[rang])
			if ip != nil {
				ips.fillIP(ip)
			} else {
				return nil, nil, errors.New("The entry is not an IP nor a range of addresses in CIDR format:" + rangs[rang])
			}
		}

	}
	// Force paralel connections
	if parallelconnections <= 0 {
		parallelconnections = paralel
	}

	ips.checkAliveHosts(port, parallelconnections, timeoutDuration)

	return ips.alive, ips.fail, nil
}
