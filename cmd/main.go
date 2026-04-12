package main

import (
	"avito-shop/internal/features/api/service"
	"avito-shop/internal/features/api/transport"
	"net/http"
)

func main() {
	var s service.Service = service.ServiceImpl{}

	r := transport.Register(s)
	http.ListenAndServe(":8080", r)
}
