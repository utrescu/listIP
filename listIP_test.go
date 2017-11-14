package listIP

import (
	"net"
	"testing"
	"time"
)

// ============= Testing fillIP() ========================================

// TestFillGoodIP : Checks if IP are added to IPList
func TestIFFillIPAcceptsGoodIPs(t *testing.T) {
	addresses := []string{"192.168.0.2", "192.168.0.3"}
	var llista IPList

	for index, address := range addresses {
		ip := net.ParseIP(address)
		llista.fillIP(ip)
		// Array size must be the same
		if len(llista.ip) != index+1 {
			t.Errorf("More IPs , %d, than expected %d: %s", len(llista.ip), index+1, llista.ip)
		}
		// Must be the same entered IP ...
		if llista.ip[index].String() != addresses[index] {
			t.Errorf("Unexpected IP: %s expected %s", llista.ip[index], addresses[index])
		}
	}
}

// TestFillBadIP checks if 'fillIP()' fails with CIDR addresses
func TestIFFillIPIgnoresBadIPs(t *testing.T) {
	addresses := []string{"192.168.0.2/28", "192.168.0.3/55", "192.168.3.2/32"}
	var llista IPList

	for _, address := range addresses {
		ip := net.ParseIP(address)
		llista.fillIP(ip)
		if ip != nil {
			t.Errorf("This is not an IP %s", address)
		}
		if len(llista.ip) != 0 {
			t.Errorf("Must not be filled : %s", llista.ip)
		}
	}
}

// ============= Testing fillNetwork() ========================================

// TestIfFillCIDRFillsCorrectIPs test the method with correct data
func TestIfFillCIDRFillsCorrectIPs(t *testing.T) {
	addresses := []string{"192.168.0.1/30", "192.168.88.0/29", "192.168.10.23/32", "192.168.9.1/31"}

	results := []int{2, 6, 1, 0}
	resultIP := []net.IP{
		net.ParseIP("192.168.0.1"), net.ParseIP("192.168.0.2"),
		net.ParseIP("192.168.88.1"), net.ParseIP("192.168.88.2"),
		net.ParseIP("192.168.88.3"), net.ParseIP("192.168.88.4"),
		net.ParseIP("192.168.88.5"), net.ParseIP("192.168.88.6"),
		net.ParseIP("192.168.10.23"),
	}

	var llista IPList

	sumIP := 0
	for index, address := range addresses {
		ip, ipnet, _ := net.ParseCIDR(address)
		llista.fillNetwork(ip, ipnet)
		sumIP = sumIP + results[index]
		if len(llista.ip) != sumIP {
			t.Errorf("Number of IP, %d, and expected are %d: %s", len(llista.ip), sumIP, llista.ip)
			break
		}
		if sumIP > len(resultIP) {
			t.Fatalf("More results %d than expected %d", sumIP, len(resultIP))
			break
		}
		for i := 0; i < len(llista.ip); i++ {
			ip1 := llista.ip[i]
			ip2 := resultIP[i]
			if ip1.String() != ip2.String() {
				t.Errorf("Received %s Expected %s", llista.ip, resultIP[:sumIP])
				break
			}

		}
	}
}

// ============= Testing checkAliveHosts() ========================================

func TestIfAliveHostsWorksWithIPs(t *testing.T) {

	listIP := [][]net.IP{
		{net.ParseIP("192.168.0.1"), net.ParseIP("192.168.0.2"),
			net.ParseIP("192.168.88.1"), net.ParseIP("192.168.88.2"),
			net.ParseIP("192.168.88.3"), net.ParseIP("192.168.88.4"),
			net.ParseIP("192.168.88.5"), net.ParseIP("192.168.88.6"),
			net.ParseIP("192.168.10.23")},
		// ----------------------
		{net.ParseIP("192.168.0.1")},
		// ----------------------
		{net.ParseIP("192.168.0.1"), net.ParseIP("192.168.0.2"),
			net.ParseIP("192.168.0.3")},
	}

	aliveOld := alive
	defer func() {
		alive = aliveOld
	}()

	alive = func(ip net.IP, port int, timeout time.Duration, messages chan string, errc chan error) {
		messages <- ip.String()
	}

	for _, oneIP := range listIP {
		var llista IPList
		llista.ip = oneIP
		llista.checkAliveHosts(22, 10, 100*time.Millisecond)

		if len(llista.alive) != len(oneIP) {
			t.Errorf("Have %d host and expect %d", len(llista.alive), len(oneIP))
		}
	}
}

// ============= Testing CHECK() ========================================

// -------------------  Test Check with incorrect dada ...
var iptests = []struct {
	networks []string
	expected int
}{
	{[]string{"192.168.0.2"}, 1},
	{[]string{"192.168.0.2/32"}, 1},
	{[]string{"192.168.0.0/24"}, 254},
	{[]string{"192.168.0.0/29"}, 6},
	{[]string{"192.168.0.1", "192.168.0.2"}, 2},
	{[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24"}, 254 * 3},
}

// TestIfCheckWorksWithCorrectData : Comprova que test retorna valors a partir de dades correctes
func TestIfCheckWorksWithCorrectData(t *testing.T) {
	aliveOld := alive
	defer func() {
		alive = aliveOld
	}()

	alive = func(ip net.IP, port int, timeout time.Duration, messages chan string, errc chan error) {
		messages <- ip.String()
	}

	for _, data := range iptests {
		result, _, err := Check(data.networks, 22, 10, "100ms")
		if err != nil {
			t.Errorf("Check must not give error with %s", data.networks)
		}
		if len(result) != data.expected {
			t.Errorf("Check gives %d results but expected %d: %s", len(result), data.expected, data.networks)
		}
	}
}

// -------------------  Test Check with incorrect data ...
var errortests = []struct {
	networks []string
	port     int
	paralel  int
	duration string
	resultat bool
}{
	{[]string{"192.168.0.2"}, 22, 1, "20ms", false},
	{[]string{"192.168.0.2/24"}, 22, 1, "20ms", false},
	{[]string{"192.168.0.2"}, 22, 1, "xxx", true},
	{[]string{"192.168.0.2"}, -1, 1, "20ms", true},
	{[]string{"192.168.0.2"}, 22, -1, "20ms", false},
	{nil, 22, 10, "20ms", true},
	{[]string{"344444444444"}, 22, 1, "20ms", true},
}

// TestIfCheckWorksWithIncorrectData checks if providing incorrect data gives error
func TestIfCheckWorksWithIncorrectData(t *testing.T) {

	for _, test := range errortests {
		_, _, err := Check(test.networks, test.port, test.paralel, test.duration)
		if test.resultat == (err == nil) {
			t.Errorf("%v not gives the expected error %t", test, test.resultat)
		}
	}

}
