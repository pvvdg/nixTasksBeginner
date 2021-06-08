package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//User ...
type User struct {
	UserId int64  `json:"userId"`
	Id     int64  `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func service(wg *sync.WaitGroup, url string) {
	user := User{}
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	json.Unmarshal(body, &user)
	fmt.Printf("{ UserId:%d Id:%d Title:%s Body:%s }\n", user.UserId, user.Id, user.Title[:10], user.Body[:10])
}

func main() {
	wg := sync.WaitGroup{}
	url := ""
	// N := 5
	N := 100
	wg.Add(N)
	for i := 1; i <= N; i++ {
		url = "https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(i)
		go service(&wg, url)
	}
	wg.Wait()
}
