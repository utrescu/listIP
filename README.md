ListIP
=====================
Go package to scan for an open port in a network.

Only exposes a single method (I'm testing Golang)

Install
---------------
Install with: 

    $ go install github.com/utrescu/ListIP

Use
-----------------
The package only exposes the method 'Comprova' to check for an open port in a list of networks

    func Check(
      rangs []string, 
      port int, 
      timeout string) []string 

Where: 

* **rangs**: list of networks in CIDR format. Ex. [192.168.0.0/24 192.168.1.1/28]
* **port**: Port to check
* **timeout**: Time to wait for mark a port as closed. Ex. "250ms" 

Example
--------------

    package main

    import (
        "fmt"
        "github.com/utrescu/listIP"
    )

    func main() {

        var rangs = []string{"192.168.1.0/24", "192.168.0.0/24"}
        portNumber := 22
        timeout := "250ms"

        results := listIP.Check(rangs, portNumber, timeout)

        fmt.Println(results)
    }

Returns a list with machines in the networks with the open port: 

   $ go run test.go 
   [192.168.1.2 192.168.0.2]