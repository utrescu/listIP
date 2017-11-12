package listIP

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

/*
IPList: Struct to work with
*/
type IPList struct {
	ip    []net.IP
	alive []string
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

func rebreResultats(outChan <-chan string, resultChan chan<- []string) {
	var alives []string

	for s := range outChan {
		alives = append(alives, s)
	}

	resultChan <- alives
}

func (n *IPList) comprovaHostsVius(port int, timeout string) {

	outChan := make(chan string, len(n.ip))
	resultChan := make(chan []string)

	numHosts := len(n.ip)

	wg.Add(numHosts)

	for i := 0; i < numHosts; i++ {
		go estaViu(n.ip[i], port, timeout, outChan)
	}

	wg.Wait()

	// Per poder fer servir el rang
	close(outChan)

	go rebreResultats(outChan, resultChan)
	n.alive = <-resultChan
}

func estaViu(ip net.IP, port int, timeout string, outChan chan<- string) {

	defer wg.Done()
	timeoutDuration, err := time.ParseDuration(timeout)

	connexio := ip.String() + ":" + strconv.Itoa(port)

	conn, err := net.DialTimeout("tcp", connexio, timeoutDuration)
	if err == nil {
		outChan <- ip.String()
		conn.Close()
	} else {
		log.Println(err.Error())
	}
	return
}

/* Determine last address */
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

	ips.comprovaHostsVius(port, timeout)

	return ips.alive
}
