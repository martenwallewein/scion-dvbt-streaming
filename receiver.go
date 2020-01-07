package main

import (
	// "context"
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
	"strconv"
	"time"
)

var laddr *snet.Addr
var err error
var local *string
var localurl *string
var remote *string
var direction *string

func ProxyToScion(wr http.ResponseWriter, r2 *http.Request) {
	c := &http.Client{
		Transport: &shttp.Transport{
			LAddr: laddr,
		},
	}
	var start time.Time
	start = time.Now()
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
	// ctx := context.Background()
	req := progress.NewReader(resp.Body)

	log.Println(size)

	/*go func() {
		progressChan := progress.NewTicker(ctx, req, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\r%v remaining...", p.Remaining().Round(time.Second))
		}
		fmt.Println("\rdownload is completed")
	}()*/

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Received status ", resp.Status)
	}

	fmt.Println("Content-Length: ", size)
	fmt.Println("Content-Type: ", resp.Header.Get("Content-Type"))

	wr.WriteHeader(200)
	_, err = io.Copy(wr, req)
	// log.Println(err)
	duration := time.Since(start)
	fmt.Printf("Total time: %v\n", duration)
	fmt.Printf("avg speed: %d bytes per ms\n", (size)/(duration.Milliseconds()))
	fmt.Println("Successfully ")

}

func ProxyFromScion(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	log.Printf("%v %v", r.Method, r.RequestURI)
	req, err = http.NewRequest(r.Method, *remote, nil)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}

	resp, err = client.Do(req)

	// combined for GET/POST
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
	defer resp.Body.Close()
}

func main() {

	local = flag.String("local", "", "The address on which the server will be listening")
	localurl = flag.String("localurl", "", "The address on which the server will be listening")
	remote = flag.String("remote", "", "The address on which the server will be requested")
	direction = flag.String("direction", "", "From ip to scion or from scion to ip")
	var port = flag.Uint("p", 40002, "port the server listens on (only relevant if local address not specified)")
	var tlsCert = flag.String("cert", "tls.pem", "Path to TLS pemfile")
	var tlsKey = flag.String("key", "tls.key", "Path to TLS keyfile")

	flag.Parse()

	if *local == "" {
		laddr, err = GetLocalhost()
	} else {
		laddr, err = snet.AddrFromString(*local)
	}
	if err != nil {
		log.Fatal(err)
	}

	if *direction == "toScion" {

		http.HandleFunc("/", ProxyToScion)
		log.Fatal(http.ListenAndServe(*localurl, nil))
	} else {
		http.HandleFunc("/", ProxyFromScion)

		var laddr string

		if *local == "" {
			laddr, err = GetLocalhostString()
			if err != nil {
				log.Fatal(err)
			}
			laddr = fmt.Sprintf("%s:%d", laddr, *port)
		} else {
			laddr = *local
		}

		log.Fatal(shttp.ListenAndServeSCION(laddr, *tlsCert, *tlsKey, nil))
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

}
