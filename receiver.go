package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/machinebox/progress"
	. "github.com/netsec-ethz/scion-apps/lib/scionutil"
	"github.com/netsec-ethz/scion-apps/lib/shttp"
	"github.com/pkg/errors"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	var local = flag.String("local", "", "The address on which the server will be listening")
	var remote = flag.String("remote", "", "The address on which the server will be requested")
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

	// InitSCION(laddr)

	// raddr, _ := snet.AddrFromString(*remote)
	// ChoosePathByMetric(MTU, laddr, raddr)
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
	resp, err := c.Get(fmt.Sprintf("https://%s:9001", *remote))
	// resp, err := c.Get("https://19-ffaa:1:c59,[127.0.0.1]:40002/image")
	if err != nil {
		log.Fatal("GET request failed: ", err)
	}
	defer resp.Body.Close()

	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		errors.New("cannot determine progress without Content-Length")
	}
	size, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	if err != nil {
		errors.Wrapf(err, "bad Content-Length %q", contentLengthHeader)
	}
	ctx := context.Background()
	r := progress.NewReader(resp.Body)

	log.Println(size)

	go func() {
		progressChan := progress.NewTicker(ctx, r, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\r%v remaining...", p.Remaining().Round(time.Second))
		}
		fmt.Println("\rdownload is completed")
	}()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Received status ", resp.Status)
	}

	fmt.Println("Content-Length: ", resp.ContentLength)
	fmt.Println("Content-Type: ", resp.Header.Get("Content-Type"))

	// bytes := 1000000
	// bytesRead := 0
	// b := make([]byte, bytes)
	/*for {
		_, err := resp.Body.Read(b)
		bytesRead += bytes
		fmt.Printf("Read %d bytes\n", bytesRead)
		if err == io.EOF {
			break
		}
	}*/
	file, err := os.Create(strings.TrimSpace("./" + "test" + ".mp4"))
	defer file.Close()
	_, err = io.Copy(file, r)
	log.Println(err)
	fmt.Println("Successfully ")
}
