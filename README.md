ListIP
=====================
Golang Package that allows you to scan if a port is open in individual IPs or in network ranges.

* Which machines have the SSH server opened in my network?
* Has 192.168.0.1 the port 80 opened?

Install
---------------
Install with:

    $ go install github.com/utrescu/ListIP

Use
-----------------
The package only exposes the method 'Check' to test for an open port:

    func Check(
      rangs []string,
      port int,
      timeout string) []string

Where:

* **rangs**: list of networks ranges in CIDR format or single IPs. Ex. [192.168.0.0/24 192.168.1.1/28 192.168.10.23]
* **port**: Port to check. Ex. 22
* **timeout**: Time to wait to receive a response. Ex. "250ms"

Returns two string lists:

1. A list of hosts with the open port
2. A list with the failed hosts

Example
--------------

    package main

    import (
        "fmt"
        "github.com/utrescu/listIP"
    )

    func main() {

        var rangs = []string{"192.168.1.0/24", "192.168.0.0/24", "192.168.9.1"}
        portNumber := 22
        timeout := "250ms"

        results, _ := listIP.Check(rangs, portNumber, timeout)

        fmt.Println(results)
    }

Returns a list of machines in the networks with the port open:

    $ go run test.go
    [192.168.1.2 192.168.0.2]
