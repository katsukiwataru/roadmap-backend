package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func handlePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/posts/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		getPost(w, r, id)
	case http.MethodPut:
		updatePost(w, r, id)
	case http.MethodDelete:
		deletePost(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getPost(w http.ResponseWriter, _ *http.Request, id int) {
	query := `SELECT id, title, content, category, tags, created_at, updated_at FROM posts WHERE id=$1`
	var post Post
	err := db.QueryRow(context.Background(), query, id).
		Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Printf("Failed to get post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getPosts(w, r)
	case http.MethodPost:
		createPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func validatePost(post Post) error {
	if post.Title == "" {
		return fmt.Errorf("Title is required")
	}
	if post.Content == "" {
		return fmt.Errorf("Content is required")
	}
	if post.Category == "" {
		return fmt.Errorf("Category is required")
	}
	return nil
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validatePost(post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO posts (title, content, category, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := db.QueryRow(context.Background(), query, post.Title, post.Content, post.Category, post.Tags).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")

	var rows pgx.Rows
	var err error

	if term != "" {
		query := `SELECT id, title, content, category, tags, created_at, updated_at
		FROM posts
		WHERE title ILIKE $1 OR content ILIKE $1 OR category ILIKE $1
		ORDER BY created_at DESC`
		rows, err = db.Query(context.Background(), query, "%"+term+"%")
	} else {
		query := `SELECT id, title, content, category, tags, created_at, updated_at FROM posts ORDER BY created_at DESC`
		rows, err = db.Query(context.Background(), query)
	}

	if err != nil {
		log.Printf("Failed to get posts: %v", err)
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan post: %v", err)
			http.Error(w, "Failed to get posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func updatePost(w http.ResponseWriter, r *http.Request, id int) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validatePost(post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE posts
		SET title=$1, content=$2, category=$3, tags=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING id, title, content, category, tags, created_at, updated_at`

	err := db.QueryRow(context.Background(), query, post.Title, post.Content, post.Category, post.Tags, id).
		Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func deletePost(w http.ResponseWriter, _ *http.Request, id int) {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
