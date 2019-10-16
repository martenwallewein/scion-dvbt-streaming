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
	"io"
	"log"
	"net/http"
)

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type HttpConnectionChannel chan *HttpConnection

var connChannel = make(HttpConnectionChannel)

func PrintHTTP(conn *HttpConnection) {
	fmt.Printf("%v %v\n", conn.Request.Method, conn.Request.RequestURI)
	for k, v := range conn.Request.Header {
		fmt.Println(k, ":", v)
	}
	fmt.Println("==============================")
	fmt.Printf("HTTP/1.1 %v\n", conn.Response.Status)
	for k, v := range conn.Response.Header {
		fmt.Println(k, ":", v)
	}
	fmt.Println(conn.Response.Body)
	fmt.Println("==============================")
}

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	//log.Printf("%v %v", r.Method, r.RequestURI)
	req, err = http.NewRequest(r.Method, "http://localhost:8008", r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	r.Body.Close()

	// combined for GET/POST
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	// conn := &HttpConnection{r, resp}

	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
	resp.Body.Close()

	// PrintHTTP(conn)
	//connChannel <- &HttpConnection{r,resp}
}

func main() {

	var local = flag.String("local", "", "The address on which the server will be listening")
	var port = flag.Uint("p", 40002, "port the server listens on (only relevant if local address not specified)")
	var tlsCert = flag.String("cert", "tls.pem", "Path to TLS pemfile")
	var tlsKey = flag.String("key", "tls.key", "Path to TLS keyfile")

	flag.Parse()
	var err error

	//simDvbt := exec.Command("cvlc", "sample.mp4", "--sout-keep", "#transcode{vcodec=x264,vb=800,scale=1,acodec=mp4a,ab=128,channels=2}:duplicate{dst=std{access=http,mux=ts,dst=localhost:8008}}")

	// simDvbt.Stdout = os.Stdout
	//err = simDvbt.Start()

	//if err != nil {
	//	println("Error setting up simulated Dvbt")
	//}

	http.HandleFunc("/image", ServeHTTP)

	/*http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		// serve the sample JPG file
		// Status 200 OK will be set implicitly
		// Content-Length will be inferred by server
		// Content-Type will be detected by server
		http.ServeFile(w, r, "./sample.mp4")
	})*/

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
