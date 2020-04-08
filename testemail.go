package main

import (
	"fmt"
	"net"
)

func mainjd() {
	mxrecords, _ := net.LookupNS("qq.com")
	for _, mx := range mxrecords {
		//fmt.Println(mx.Host, mx.Pref)
		fmt.Println(mx)
	}
	fmt.Println(mxrecords)
	cname, err := net.LookupMX("163.com")
	if err != nil {
		panic(err)
	}
	// dig +short research.swtch.com cname
	for _, mx := range cname {
		fmt.Println(mx.Host, mx.Pref)
	}
}
