package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type Posts []Post

type Post struct {
	UserId int64  `json:"userId"`
	Id     int64  `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Comment struct {
	PostId int64  `json:"postId"`
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

func getPosts(url string) *Posts {
	var posts = &Posts{}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	json.Unmarshal(body, &posts)

	return posts
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

func (p *Posts) showDBPosts(db *sql.DB, nameDB string) {
	rows, err := db.Query("SELECT * FROM " + nameDB)
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

func (p *Posts) insertDBPosts(db *sql.DB, nameDB string) {
	stmt, err := db.Prepare("INSERT " + nameDB + " SET user_id=?,id=?,title=?,body=?")
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

func main() {
	db := dbConn()
	defer db.Close()

	var version string
	errQueryRow := db.QueryRow("SELECT VERSION()").Scan(&version)
	if errQueryRow != nil {
		log.Fatal(errQueryRow)
	}
	log.Println("Connection:OK version MySQL is", version)

	var wg sync.WaitGroup
	wg.Add(2)
	var posts *Posts
	// nameDB := "posts"
	url := "https://jsonplaceholder.typicode.com/posts?userId=7"
	posts = getPosts(url)
	N := len(*posts)
	fmt.Println(N)
	// go posts.insertDBPosts(db, nameDB)
	// go posts.showDBPosts(db, nameDB)

}
