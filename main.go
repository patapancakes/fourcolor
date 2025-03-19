/*
	fourcolor - a simple anonymous bbs
	Copyright (C) 2025  Pancakes <patapancakes@pagefault.games>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"fourcolor/db"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/xeonx/timeago"
)

type TemplateData struct {
	Site   SiteConfig
	Posts  []db.Post
	ViewID int
	Ticket string
}

type SiteConfig struct {
	Name   string
	Slogan string
}

var (
	t      = template.Must(template.New("base.html").Funcs(template.FuncMap{"niceauthor": base64.StdEncoding.EncodeToString, "nicetime": timeago.English.Format}).ParseGlob("data/templates/*"))
	config = SiteConfig{
		Name:   "FourColor BBS",
		Slogan: "Post away... or don't, it's up to you.",
	}
)

func main() {
	// flags
	dbuser := flag.String("dbuser", "fourcolor", "database user's name")
	dbpass := flag.String("dbpass", "", "database user's password")
	dbaddr := flag.String("dbaddr", "localhost", "database server address")
	dbname := flag.String("dbname", "fourcolor", "database name")
	port := flag.Int("port", 80, "port to listen on")
	flag.Parse()

	// database
	err := db.Init(*dbuser, *dbpass, *dbaddr, *dbname)
	if err != nil {
		log.Fatal(err)
	}

	authorKey, err = getAuthorKey()
	if err != nil {
		log.Fatal(err)
	}

	ticketKey, err = getTicketKey()
	if err != nil {
		log.Fatal(err)
	}

	// set up http routes
	http.HandleFunc("GET /", handleHome)
	http.HandleFunc("GET /{id}/", handleThread)
	http.HandleFunc("POST /post", handlePost)

	http.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("data/assets"))))

	// start http server
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	posts, err := db.FetchPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// newest first
	slices.Reverse(posts)

	ticket := base64.StdEncoding.EncodeToString(deriveTicket(r.Header.Get("X-Forwarded-For"), r.UserAgent()))

	err = t.Execute(w, TemplateData{Site: config, Posts: posts, Ticket: ticket})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleThread(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := db.FetchPost(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if post.IsReply() {
		http.Redirect(w, r, fmt.Sprintf("/%d/#post_%d", post.Parent.Int64, id), http.StatusSeeOther)
		return
	}

	replies, err := db.FetchPostsWithParent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticket := base64.StdEncoding.EncodeToString(deriveTicket(r.Header.Get("X-Forwarded-For"), r.UserAgent()))

	err = t.Execute(w, TemplateData{Site: config, Posts: slices.Insert(replies, 0, post), Ticket: ticket, ViewID: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := strings.TrimSpace(r.Form.Get("comment"))

	if len(body) == 0 {
		http.Error(w, "missing message body", http.StatusBadRequest)
		return
	}

	host := r.Header.Get("X-Forwarded-For")

	if r.Form.Get("ticket") != base64.StdEncoding.EncodeToString(deriveTicket(host, r.UserAgent())) {
		http.Error(w, "invalid ticket", http.StatusBadRequest)
		return
	}

	if r.Form.Has("parent") {
		parent, err := strconv.Atoi(r.Form.Get("parent"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := db.InsertPostWithParent(parent, deriveAuthorDigest(host, parent), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%d/#post_%d", parent, id), http.StatusSeeOther)
		return
	}

	id, err := db.InsertPost(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.UpdatePostAuthor(deriveAuthorDigest(host, int(id)), int(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%d/", id), http.StatusSeeOther)
}
