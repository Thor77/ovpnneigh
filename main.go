package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
)

const ipp = "/etc/openvpn/server/ipp.txt"
const networkInterface = "ens3"

// thanks https://stackoverflow.com/a/10485970/2683991
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func parseIPPFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var addresses []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addresses = append(addresses, strings.Split(scanner.Text(), ",")[2])
	}
	return addresses, nil
}

func main() {
	// read ipp.txt for client ip addresses
	requiredProxyAddresses, err := parseIPPFile(ipp)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d addresses: %v\n", len(requiredProxyAddresses), requiredProxyAddresses)

	// get already set proxy addresses
	var presentProxyAddresses []string
	output, err := exec.Command("/usr/bin/ip", "-6", "neigh", "list", "proxy").Output()
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		presentProxyAddresses = append(presentProxyAddresses, strings.Split(line, " ")[0])
	}
	log.Printf("Proxy is already active for %d addresses: %v\n", len(presentProxyAddresses), presentProxyAddresses)

	var missingProxyAddresses []string
	for _, address := range requiredProxyAddresses {
		if !contains(presentProxyAddresses, address) {
			missingProxyAddresses = append(missingProxyAddresses, address)
		}
	}

	log.Printf("%d address left to add: %v\n", len(missingProxyAddresses), missingProxyAddresses)
	for _, address := range missingProxyAddresses {
		// add proxy for address
		cmd := exec.Command("/usr/bin/ip", "-6", "neigh", "add", "proxy", address, "dev", networkInterface)
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
