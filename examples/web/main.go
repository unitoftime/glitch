package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Serving on 8081")
	http.ListenAndServe(`:8081`, http.FileServer(http.Dir(`.`)))
}
