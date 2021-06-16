package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Posts []Post

type Post struct {
	UserId int    `json:"userId"`
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Comments []Comment

type Comment struct {
	PostId int    `json:"postId"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "desgrad"
	dbName := "jsonplaceholder"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func checkConnectAndVersionDB(db *sql.DB) {
	var version string
	errQueryRow := db.QueryRow("SELECT VERSION()").Scan(&version)
	if errQueryRow != nil {
		log.Fatal(errQueryRow)
	}
	log.Println("Connection:OK version MySQL is", version)
}

func getBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return body, err
}

func getPosts(url string) *Posts {
	var posts = &Posts{}
	body, err := getBody(url)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &posts)
	return posts
}

func getAndAppendCommentsIntoDB(db *sql.DB, nameDB string, chanPostID chan int, countPosts int) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	var comments = Comments{}
	url := "https://jsonplaceholder.typicode.com/comments?postId="
	wg.Add(countPosts)
	for i := 0; i < countPosts; i++ {
		go func() {
			mu.Lock()
			defer wg.Done()
			if val, opened := <-chanPostID; opened {
				body, err := getBody(url + strconv.Itoa(val))
				if err != nil {
					log.Fatal(err)
				}
				json.Unmarshal(body, &comments)
				countComments := len(comments)
				for j := 0; j < countComments; j++ {
					go func(j int) {
						if _, err := db.Exec("INSERT INTO "+nameDB+" (`post_id`,`id`,`name`,`email`,`body`) VALUES(?,?,?,?,?)",
							comments[j].PostId, comments[j].Id, comments[j].Name, comments[j].Email, comments[j].Body); err != nil {
							log.Fatal(err)
						}
					}(j)
				}
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
}

func insertIntoDBPosts(db *sql.DB, nameDB string) {
	var posts = Posts{}
	stmt, err := db.Prepare("INSERT " + nameDB + " SET user_id=?,id=?,title=?,body=?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, v := range posts {
		if _, err := stmt.Exec(v.UserId, v.Id, v.Title, v.Body); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Insert OK")
}

func main() {
	start := time.Now()

	db := dbConn()

	checkConnectAndVersionDB(db)

	url := "https://jsonplaceholder.typicode.com/posts?userId=7"
	posts := getPosts(url)
	countPosts := len(*posts)
	nameDB := "posts"

	go insertIntoDBPosts(db, nameDB)

	chanPostID := make(chan int, countPosts)

	for _, v := range *posts {
		chanPostID <- v.Id
	}

	nameDB = "comments"
	getAndAppendCommentsIntoDB(db, nameDB, chanPostID, countPosts)
	fmt.Println(time.Since(start))

	db.Close()
}
