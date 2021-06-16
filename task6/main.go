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

func getAndAppendCommentsIntoDB(chanPostID chan int, N int) {
	start := time.Now()
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	var comments = &Comments{}
	url := "https://jsonplaceholder.typicode.com/comments?postId="
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			mu.Lock()
			defer wg.Done()
			if val, opened := <-chanPostID; opened {
				body, err := getBody(url + strconv.Itoa(val))
				if err != nil {
					log.Fatal(err)
				}
				json.Unmarshal(body, comments)

				for _, v := range *comments {
					go func(v Comment) {
						fmt.Println(v.Id, "\t", v.Email, "\t ")
					}(v)
				}
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(time.Since(start))
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

func (c *Comments) showFromDBComments(db *sql.DB, nameDB string) {
	rows, err := db.Query("SELECT * FROM " + nameDB + ";")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var postId int
		var id int
		var name string
		var email string
		var body string
		err = rows.Scan(&postId, &id, &name, &email, &body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v | %v | %v | %v | %v\n", postId, id, name[:10], email, body[:10])
	}
}

func insertIntoDBComments(db *sql.DB, nameDB string) {
	c := &Comments{}
	stmt, err := db.Prepare("INSERT " + nameDB + " SET user_id=?,id=?,title=?,body=?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, v := range *c {
		if _, err := stmt.Exec(v.PostId, v.Id, v.Name, v.Email, v.Body); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Insert OK")
}

func (p *Posts) showFromDBPosts(db *sql.DB, nameDB string) {
	rows, err := db.Query("SELECT * FROM " + nameDB + ";")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var user_id int
		var id int
		var title string
		var body string
		err = rows.Scan(&user_id, &id, &title, &body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v | %v | %v | %v |\n", user_id, id, title[:10], body[:10])
	}
}

func (p *Posts) insertIntoDBPosts(db *sql.DB, nameDB string) {
	stmt, err := db.Prepare("INSERT " + nameDB + " SET user_id=?,id=?,title=?,body=?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, v := range *p {
		if _, err := stmt.Exec(v.UserId, v.Id, v.Title, v.Body); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Insert OK")
}

func (p *Posts) deleteFromDBPosts(db *sql.DB, nameDB string) {
	stmt, err := db.Prepare("DELETE FROM " + nameDB + " WHERE id=?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, v := range *p {
		if _, err := stmt.Exec(v.Id); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Delete OK")
}

func main() {
	db := dbConn()
	defer db.Close()

	checkConnectAndVersionDB(db)

	url := "https://jsonplaceholder.typicode.com/posts?userId=7"
	posts := getPosts(url)
	countPosts := len(*posts)

	chanPostID := make(chan int, countPosts)

	for _, v := range *posts {
		chanPostID <- v.Id
	}

	getAndAppendCommentsIntoDB(chanPostID, countPosts)
	// fmt.Println(comments)

	/*
		nameDB := "posts"
		posts.insertDBPosts(db, nameDB)
		posts.showDBPosts(db, nameDB)
		posts.deleteDBPosts(db, nameDB)
		posts.showDBPosts(db, nameDB)
	*/

}
