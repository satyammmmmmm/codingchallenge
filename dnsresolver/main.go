package main

import (
	"fmt"
	"os"

	"github.com/satyammmmmmm/codingchallenge/tree/main/dnsresolver/resolver"
)

func main() {
	searchHostname := os.Args[1]

	rootnameServer := "198.41.0.4"
	ipAddress, err := resolver.DomainnameResolver(searchHostname, rootnameServer)
	if err != nil {
		fmt.Printf("error in resolving domainname", err)
		os.Exit(1)
	}
	fmt.Println(ipAddress)

}
