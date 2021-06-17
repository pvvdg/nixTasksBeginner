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
	var p = &Posts{}
	body, err := getBody(url)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &p)
	return p
}

/*
func getComments(url string) *Comments {
	var c = &Comments{}
	body, err := getBody(url)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &c)
	return c
}*/

func getAndAppendCommentsIntoDB(wg *sync.WaitGroup, db *sql.DB, nameDB string, chanPostID chan int, countPosts int) {
	stmt := `INSERT INTO ` + nameDB + ` (post_id,id,name,email,body) VALUES(?,?,?,?,?)`
	var c = &Comments{}
	url := "https://jsonplaceholder.typicode.com/comments?postId="
	wg.Add(countPosts)
	for i := 0; i < countPosts; i++ {
		go func() {
			defer wg.Done()
			if val, opened := <-chanPostID; opened {
				body, err := getBody(url + strconv.Itoa(val))
				if err != nil {
					log.Fatal(err)
				}
				json.Unmarshal(body, &c)

				countComments := len(*c)
				wg.Add(countComments)
				for i, v := range *c {
					v := v
					go func(i int, v *Comment) {
						defer wg.Done()
						if _, err := db.Exec(stmt, v.PostId, v.Id, v.Name, v.Email, v.Body); err != nil {
							log.Fatal(err)
						}
						fmt.Println(i, " ", v.Id)
					}(i, &v)
				}
				// wg.Wait()
			}
		}()
	}
	wg.Wait()
}

func (p *Posts) insertIntoDBPosts(wg *sync.WaitGroup, db *sql.DB, nameDB string) {
	stmt := `INSERT INTO ` + nameDB + ` (user_id,id,title,body) VALUES(?,?,?,?)`
	lenPosts := len(*p)
	wg.Add(lenPosts)
	for _, v := range *p {
		v := v
		go func(v *Post) {
			defer wg.Done()
			if _, err := db.Exec(stmt, v.UserId, v.Id, v.Title, v.Body); err != nil {
				log.Fatal(err)
			}
		}(&v)
	}
	wg.Wait()
}

func main() {
	start := time.Now()

	db := dbConn()
	defer db.Close()

	checkConnectAndVersionDB(db)

	wg := &sync.WaitGroup{}

	url := "https://jsonplaceholder.typicode.com/posts?userId=7"
	posts := getPosts(url)
	countPosts := len(*posts)
	nameDB := "posts"

	posts.insertIntoDBPosts(wg, db, nameDB)

	chanPostID := make(chan int, countPosts)

	for _, v := range *posts {
		chanPostID <- v.Id
	}

	nameDB = "comments"
	getAndAppendCommentsIntoDB(wg, db, nameDB, chanPostID, countPosts)

	fmt.Println(time.Since(start))
}
