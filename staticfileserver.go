package main

import (
	"net/http"
)

func fileh(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./1G.file")
}

func main() {
	http.HandleFunc("/", fileh)
	panic(http.ListenAndServe("141.44.25.148:8088", nil))
}
