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
	"strconv"
	"time"
)

var local *string
var remote *string
var remoteIp *string
var direction *string
var err error
var laddr string
var lsaddr *snet.Addr

func ProxyToScion(wr http.ResponseWriter, r2 *http.Request) {
	c := &http.Client{
		Transport: &shttp.Transport{
			LAddr: lsaddr,
		},
	}

	// Make a get request
	resp, err := c.Get(fmt.Sprintf("http//%s:9001", *remote))
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
	req := progress.NewReader(resp.Body)

	log.Println(size)

	go func() {
		progressChan := progress.NewTicker(ctx, req, size, 1*time.Second)
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

	wr.WriteHeader(200)
	_, err = io.Copy(wr, req)
	log.Println(err)
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
	remote = flag.String("remote", "", "The address on which the server will be sebd")
	remoteIp = flag.String("remoteip", "", "The scion address on which the server will be sebd")
	direction = flag.String("direction", "", "From ip to scion or from scion to ip")
	var port = flag.Uint("p", 40002, "port the server listens on (only relevant if local address not specified)")
	var tlsCert = flag.String("cert", "tls.pem", "Path to TLS pemfile")
	var tlsKey = flag.String("key", "tls.key", "Path to TLS keyfile")
	var localUrl = flag.String("localurl", "localhost:8008", "Ip address of the current server")

	flag.Parse()

	if *direction == "toScion" {

		if *local == "" {
			lsaddr, err = GetLocalhost()
		} else {
			lsaddr, err = snet.AddrFromString(*local)
		}
		if err != nil {
			log.Fatal(err)
		}

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
