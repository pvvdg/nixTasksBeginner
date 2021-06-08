package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	URL := "https://jsonplaceholder.typicode.com/posts"
	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	fmt.Println(sb)
}
