package main

import (
	"net/http"
)

func fileh(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/sample.mp4")
}

func main() {
	http.HandleFunc("/", fileh)
	panic(http.ListenAndServe("78.47.172.41:8088", nil))
}
