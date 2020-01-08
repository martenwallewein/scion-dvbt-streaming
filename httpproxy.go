package main

import (
	// "context"
	"flag"
	"fmt"
	"github.com/machinebox/progress"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var err error
var localurl *string
var remote *string

func Proxy(wr http.ResponseWriter, r2 *http.Request) {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	var start time.Time
	start = time.Now()
	// Make a get request
	resp, err := c.Get(fmt.Sprintf("http://%s", *remote))
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
	durMillies := duration.Milliseconds()
	if durMillies > 0 {
		fmt.Printf("avg speed: %d bytes per ms\n", (size)/(durMillies))
	}

	fmt.Println("Successfully ")

}

func main() {

	localurl = flag.String("localurl", "", "The address on which the server will be listening")
	remote = flag.String("remote", "", "The address on which the server will be requested")

	flag.Parse()
	http.HandleFunc("/", Proxy)
	log.Fatal(http.ListenAndServe(*localurl, nil))
}
