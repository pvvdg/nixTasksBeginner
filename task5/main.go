package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func requestToFile(wg *sync.WaitGroup, url string, i int) {
	defer wg.Done()
	user := User{}
	url += strconv.Itoa(i)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	json.Unmarshal(body, &user)
	file, err := os.Create(strconv.Itoa(i) + ".txt")
	if err != nil {
		log.Fatalln("Unable to create file:", err)
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("{ UserId:%d Id:%d Title:%s Body:%s }\n", user.UserId, user.Id, user.Title[:10], user.Body[:10]))
}

func main() {
	wg := sync.WaitGroup{}
	url := "https://jsonplaceholder.typicode.com/posts/"
	N := 5
	// N := 100
	wg.Add(N)
	for i := 1; i <= N; i++ {
		go requestToFile(&wg, url, i)
	}
	wg.Wait()
	log.Println("OK")
}
