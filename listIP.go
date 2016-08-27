package listIP

import (
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

/*
  fill: fills all network adresses
*/
func (n *IPList) fill(ip net.IP, ipnet *net.IPNet) {
	notfirst := false
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {

		if !notfirst {
			// remove network address
			notfirst = true
			continue
		}
		novaIP := make(net.IP, len(ip))
		copy(novaIP, ip)

		n.ip = append(n.ip, novaIP)

	}
	// Remove broadcast if any ...
	if len(n.ip) > 0 {
		n.ip = n.ip[0 : len(n.ip)-1]
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
	}
	return
}

/*
Check a list of CIDR networks for the specified port open
*/
func Check(rangs []string, port int, timeout string) []string {
	var ips IPList

	for rang := range rangs {
		ip, ipnet, err := net.ParseCIDR(rangs[rang])
		if err != nil {
			log.Fatal(err)
		}
		ips.fill(ip, ipnet)
	}

	ips.comprovaHostsVius(port, timeout)

	return ips.alive
}
