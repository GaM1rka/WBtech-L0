package main

import (
	"net/http"
	"readermicroservice/configs"
	"readermicroservice/internal"
)

func main() {
	configs.Configure()
	http.HandleFunc("/order/", internal.OrderHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return
	}
}
