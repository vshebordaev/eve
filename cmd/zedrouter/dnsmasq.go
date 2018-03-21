// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

// dnsmasq configlets for overlay and underlay interfaces towards domU

package main

import (
	"fmt"
	"github.com/zededa/go-provision/wrap"
	"log"
	"os"
)

const dnsmasqOverlayStatic = `
# Automatically generated by zedrouter
except-interface=lo
bind-interfaces
log-queries
log-dhcp
no-hosts
no-ping
bogus-priv
stop-dns-rebind
rebind-localhost-ok
neg-ttl=10
dhcp-range=::,static,0,infinite
`

const dnsmasqUnderlayStatic = `
# Automatically generated by zedrouter
except-interface=lo
bind-interfaces
log-queries
log-dhcp
no-hosts
no-ping
bogus-priv
stop-dns-rebind
rebind-localhost-ok
neg-ttl=10
dhcp-range=172.27.0.0,static,255.255.0.0,infinite
`

// Create the dnsmasq configuration for the the overlay interface
// Would be more polite to return an error then to Fatal
func createDnsmasqOverlayConfiglet(cfgPathname string, olIfname string,
	olAddr1 string, olAddr2 string, olMac string, hostsDir string,
	hostName string, ipsets []string) {
	if debug {
		log.Printf("createDnsmasqOverlayConfiglen: %s\n", olIfname)
	}
	file, err := os.Create(cfgPathname)
	if err != nil {
		log.Fatal("os.Create for ", cfgPathname, err)
	}
	defer file.Close()
	file.WriteString(dnsmasqOverlayStatic)
	for _, ipset := range ipsets {
		file.WriteString(fmt.Sprintf("ipset=/%s/ipv4.%s,ipv6.%s\n",
			ipset, ipset, ipset))
	}
	file.WriteString(fmt.Sprintf("pid-file=/var/run/dnsmasq.%s.pid\n",
		olIfname))
	file.WriteString(fmt.Sprintf("interface=%s\n", olIfname))
	file.WriteString(fmt.Sprintf("listen-address=%s\n", olAddr1))
	file.WriteString(fmt.Sprintf("dhcp-host=%s,[%s],%s\n",
		olMac, olAddr2, hostName))
	file.WriteString(fmt.Sprintf("hostsdir=%s\n", hostsDir))
}

// Create the dnsmasq configuration for the the underlay interface
// Would be more polite to return an error then to Fatal
func createDnsmasqUnderlayConfiglet(cfgPathname string, ulIfname string,
	ulAddr1 string, ulAddr2 string, ulMac string, hostName string, ipsets []string) {
	if debug {
		log.Printf("createDnsmasqUnderlayConfiglen: %s\n", ulIfname)
	}
	file, err := os.Create(cfgPathname)
	if err != nil {
		log.Fatal("os.Create for ", cfgPathname, err)
	}
	defer file.Close()
	file.WriteString(dnsmasqUnderlayStatic)
	for _, ipset := range ipsets {
		file.WriteString(fmt.Sprintf("ipset=/%s/ipv4.%s,ipv6.%s\n",
			ipset, ipset, ipset))
	}
	file.WriteString(fmt.Sprintf("pid-file=/var/run/dnsmasq.%s.pid\n",
		ulIfname))
	file.WriteString(fmt.Sprintf("interface=%s\n", ulIfname))
	file.WriteString(fmt.Sprintf("listen-address=%s\n", ulAddr1))
	file.WriteString(fmt.Sprintf("dhcp-host=%s,id:*,%s,%s\n",
		ulMac, ulAddr2, hostName))
}

func deleteDnsmasqConfiglet(cfgPathname string) {
	if debug {
		log.Printf("deleteDnsmasqOverlayConfiglen: %s\n", cfgPathname)
	}
	if err := os.Remove(cfgPathname); err != nil {
		log.Println(err)
	}
}

// Run this:
//    DMDIR=/opt/zededa/bin/
//    ${DMDIR}/dnsmasq --conf-file=/var/run/zedrouter/dnsmasq.${OLIFNAME}.conf
// or
//    ${DMDIR}/dnsmasq --conf-file=/var/run/zedrouter/dnsmasq.${ULIFNAME}.conf
func startDnsmasq(cfgPathname string) {
	if debug {
		log.Printf("startDnsmasq: %s\n", cfgPathname)
	}
	cmd := "nohup"
	args := []string{
		"/opt/zededa/bin/dnsmasq",
		"-C",
		cfgPathname,
	}
	go wrap.Command(cmd, args...).Output()
}

//    pkill -u nobody -f dnsmasq.${IFNAME}.conf
func stopDnsmasq(cfgFilename string, printOnError bool) {
	if debug {
		log.Printf("stopDnsmasq: %s\n", cfgFilename)
	}
	pkillUserArgs("nobody", cfgFilename, printOnError)
}
