package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Test struct {
	id        int
	member_id string
}

type Config struct {
	Database DatabaseConfig `db:database`
}

type DatabaseConfig struct {
	Host     string `db:"host"`
	Port     int    `db:"port"`
	Username string `db:username`
	Password string `db:password`
	Dbname   string `db:dbname`
}

type Post struct {
	ID   uint   `db:"id"`
	Text string `db:"text"`
	// img_file_name string
	// star_count    int
}

type Artist struct {
	ID   uint
	Name string
}

type StarCount struct {
	StarCount int `json:star_count`
}

var tmpl *template.Template

func main() {
	tmpl = template.Must(template.ParseGlob("./views/*.html"))
	m := martini.Classic()
	m.Use(martini.Static("public"))
	db, err := sqlx.Open("sqlite3", "./db/post.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	m.Get("/", func(res http.ResponseWriter, req *http.Request) {
		posts := []Post{}
		err := db.Select(&posts, "SELECT id, text FROM posts ORDER BY id DESC")
		fmt.Printf("%+v\n", posts)
		if err != nil {
			log.Fatal(err)
		}

		tmpl.ExecuteTemplate(res, "index.html", map[string]interface{}{
			"posts": posts,
		})
	})

	m.Post("/", func(res http.ResponseWriter, req *http.Request) {
		if err != nil {
			log.Fatal(err)
		}
		body, _ := ioutil.ReadAll(req.Body)
		params, _ := url.ParseQuery(string(body))
		fmt.Printf("%+v\n", params["ex_text"][0])
		db.Exec(`INSERT INTO posts (text) VALUES (?)`, params["ex_text"][0])
	})

	m.Get("/star", func(res http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		if params["post_id"] != nil {
			postID := params["post_id"][0]
			var starCnt int
			err := db.Get(&starCnt, "SELECT star_count FROM posts WHERE id = ?", postID)
			if err != nil {
				log.Fatal(err)
			}
			newStarcnt := starCnt + 1
			db.Exec("UPDATE posts SET star_count = ? WHERE id = ?", newStarcnt, postID)
			jsonBytes, jsonErr := json.Marshal(StarCount{StarCount: newStarcnt})
			if jsonErr != nil {
				res.WriteHeader(http.StatusInternalServerError)
			}
			body := string(jsonBytes)
			headers := make(map[string]string)
			headers["Content-Type"] = "application/json"
			headers["Content-Length"] = strconv.Itoa(len(body))
			for name, value := range headers {
				res.Header().Set(name, value)
			}
			res.WriteHeader(http.StatusOK)
			io.WriteString(res, body)
		} else {
		}
	})

	m.Run()
}
