// Copyright 2018 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package main

// https://stackoverflow.com/questions/34855343/how-to-stream-an-io-readcloser-into-a-http-responsewriter-in-golang

package main

import (
	"flag"
	"fmt"
	. "github.com/netsec-ethz/scion-apps/lib/scionutil"
	"github.com/netsec-ethz/scion-apps/lib/shttp"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type HttpConnectionChannel chan *HttpConnection

var connChannel = make(HttpConnectionChannel)
var local *string
var remote *string
var remoteIp *string
var direction *string

func PrintHTTP(conn *HttpConnection) {
	log.Printf("%v %v\n", conn.Request.Method, conn.Request.RequestURI)
	for k, v := range conn.Request.Header {
		fmt.Println(k, ":", v)
	}
	log.Println("==============================")
	log.Printf("HTTP/1.1 %v\n", conn.Response.Status)
	for k, v := range conn.Response.Header {
		fmt.Println(k, ":", v)
	}
	log.Println(conn.Response.Body)
	log.Println("==============================")
}

func ProxyToScion(wr http.ResponseWriter, r *http.Request) {

	var err error
	// var req *http.Request

	var laddr *snet.Addr

	if *local == "" {
		laddr, err = GetLocalhost()
	} else {
		laddr, err = snet.AddrFromString(*local)
	}
	if err != nil {
		log.Fatal(err)
	}

	raddr, err := snet.AddrFromString(*remoteIp)
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

	ChoosePathByMetric(Shortest, laddr, raddr)

	// Create a standard server with our custom RoundTripper
	c := &http.Client{
		Transport: &shttp.Transport{
			LAddr: laddr,
		},
		Timeout: 5 * time.Second,
	}

	// Make a get request

	req, err := http.NewRequest(r.Method, fmt.Sprintf("https://%s", *remote), nil)
	resp, err := c.Do(req)
	// resp, err := c.Get("https://19-ffaa:1:c59,[127.0.0.1]:40002/image")
	if err != nil {
		log.Fatal("GET request failed: ", err)
	}

	// if resp.StatusCode != http.StatusOK {
	// 	log.Fatal("Received status ", resp.Status)
	// }

	// fmt.Println("Content-Length: ", resp.ContentLength)
	// fmt.Println("Content-Type: ", resp.Header.Get("Content-Type"))

	// log.Printf("Request proxied, reading response...")

	// conn := &HttpConnection{r, resp}
	// PrintHTTP(conn)
	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}
	wr.WriteHeader(resp.StatusCode)

	file, err := os.Create(strings.TrimSpace("./" + "test" + ".mp4"))
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	log.Println(err)

	// log.Println("Copied to local file")

	//_, err = io.Copy(wr, resp.Body)
	// log.Println(err)

	log.Println("Copied to request")

	defer resp.Body.Close()

	// defer resp.Body.Close()

	//connChannel <- &HttpConnection{r,resp}
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
	// req.Close = true
	resp, err = client.Do(req)
	// r.Body.Close()

	// combined for GET/POST
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	// conn := &HttpConnection{r, resp}

	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}

	// PrintHTTP(conn)

	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
	defer resp.Body.Close()

	//connChannel <- &HttpConnection{r,resp}

	// log.Println("Serving sample.mp4")
	// http.ServeFile(wr, r, "./sample.mp4")
}

func main() {

	local = flag.String("local", "", "The address on which the server will be listening")
	remote = flag.String("remote", "", "The address on which the server will be sebd")
	remoteIp = flag.String("remoteip", "", "The scion address on which the server will be sebd")
	direction = flag.String("direction", "", "From ip to scion or from scion to ip")
	var port = flag.Uint("p", 40002, "port the server listens on (only relevant if local address not specified)")
	var tlsCert = flag.String("cert", "tls.pem", "Path to TLS pemfile")
	var tlsKey = flag.String("key", "tls.key", "Path to TLS keyfile")
	var localUrl = flag.String("localurl", "localhost:8008", "Ip address of the current server")

	flag.Parse()
	var err error

	if *direction == "toScion" {
		http.HandleFunc("/", ProxyToScion)
		log.Fatal(http.ListenAndServe(*localUrl, nil))
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

}

// ./scionhttpproxy --local=17-ffaa:1:cf1,[78.46.129.90]:9001 --remote=http://78.46.129.90:8088 --direction=fromScion --cert tls.pem --key tls.key
// ./scionhttpproxy --remote=17-ffaa:1:cf1,[78.46.129.90]:9001 --localurl=78.46.129.90:9002 --direction=toScion

// ./scionhttpproxy --local="19-ffaa:1:c3f,[141.44.25.148]:9001" --remote="http://78.46.129.90:8088" --direction=fromScion --cert tls.cert --key tls.key

// ./scionhttpproxy --remote="19-ffaa:1:c3f,[141.44.25.148]:9001" --localurl="141.44.25.151:9002" --direction=toScion
