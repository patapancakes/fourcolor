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

package db

import (
	"database/sql"
	"time"
)

type Post struct {
	ID     int
	Parent sql.NullInt64
	Author []byte
	Body   string
	Time   time.Time
}

func (p Post) IsThread() bool {
	return !p.Parent.Valid
}

func (p Post) IsReply() bool {
	return p.Parent.Valid
}

func FetchPost(id int) (Post, error) {
	var post Post
	err := handle.QueryRow("SELECT id, parent, author, body, posted FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Parent, &post.Author, &post.Body, &post.Time)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func FetchPosts() ([]Post, error) {
	r, err := handle.Query("SELECT id, parent, author, body, posted FROM posts WHERE parent IS NULL")
	if err != nil {
		return nil, err
	}

	var posts []Post
	for r.Next() {
		var post Post
		err = r.Scan(&post.ID, &post.Parent, &post.Author, &post.Body, &post.Time)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func FetchPostsWithParent(parent int) ([]Post, error) {
	r, err := handle.Query("SELECT id, parent, author, body, posted FROM posts WHERE parent = ?", parent)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for r.Next() {
		var post Post
		err = r.Scan(&post.ID, &post.Parent, &post.Author, &post.Body, &post.Time)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func InsertPost(body string) (int64, error) {
	r, err := handle.Exec("INSERT INTO posts (body) VALUES (?)", body)
	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func InsertPostWithParent(parent int, author []byte, body string) (int64, error) {
	r, err := handle.Exec("INSERT INTO posts (parent, author, body) VALUES (?, ?, ?)", parent, author, body)
	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdatePostAuthor(author []byte, id int) error {
	_, err := handle.Exec("UPDATE posts SET author = ? WHERE id = ?", author, id)
	if err != nil {
		return err
	}

	return nil
}
