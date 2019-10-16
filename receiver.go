package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	. "github.com/netsec-ethz/scion-apps/lib/scionutil"
	"github.com/netsec-ethz/scion-apps/lib/shttp"
	"github.com/scionproto/scion/go/lib/snet"
)

func main() {

	var local = flag.String("local", "", "The address on which the server will be listening")
	// var interactive = flag.Bool("i", false, "Wether to use interactive mode for path selection")

	flag.Parse()

	var laddr *snet.Addr
	var err error
	if *local == "" {
		laddr, err = GetLocalhost()
	} else {
		laddr, err = snet.AddrFromString(*local)
	}
	if err != nil {
		log.Fatal(err)
	}

	/*ia, l3, err := GetHostByName("image-server")
	if err != nil {
		log.Fatal(err)
	}
	l4 := addr.NewL4UDPInfo(40002)
	raddr := &snet.Addr{IA: ia, Host: &addr.AppAddr{L3: l3, L4: l4}}

	if *interactive {
		ChoosePathInteractive(laddr, raddr)
	} else {
		ChoosePathByMetric(Shortest, laddr, raddr)
	}*/

	// Create a standard server with our custom RoundTripper
	c := &http.Client{
		Transport: &shttp.Transport{
			LAddr: laddr,
		},
	}

	// Make a get request
	resp, err := c.Get("https://19-ffaa:1:c59,[127.0.0.1]:40002/image")
	if err != nil {
		log.Fatal("GET request failed: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Received status ", resp.Status)
	}

	fmt.Println("Content-Length: ", resp.ContentLength)
	fmt.Println("Content-Type: ", resp.Header.Get("Content-Type"))

	bytes := 1024 * 1024

	b := make([]byte, bytes)
	for {
		_, err := resp.Body.Read(b)
		fmt.Printf("Read %d bytes", bytes)
		if err == io.EOF {
			break
		}
	}

	fmt.Println("Successfully ")
}
