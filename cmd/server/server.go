package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const storagePath = "/tmp/knoxite.storage"

func authPath(w http.ResponseWriter, r *http.Request) (string, error) {
	auth, _, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: no auth set")
	}

	// check for relative path attacks
	if strings.Contains(r.URL.Path, ".."+string(os.PathSeparator)) {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: url path tampering")
	}
	if strings.Contains(auth, ".."+string(os.PathSeparator)) {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: auth code tampering")
	}

	dir := filepath.Join(storagePath, auth)
	src, err := os.Stat(dir)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Invalid auth code: unknown user")
	}
	if !src.IsDir() {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Invalid auth code: not a dir")
	}

	return dir, nil
}

// upload logic.
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving upload")
	if r.Method == "POST" {
		path, err := authPath(w, r)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}

		err = r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile(filepath.Join(path, "chunks", handler.Filename), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("Stored chunk", filepath.Join(path, "chunks", handler.Filename))
	}
}

// download logic.
func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving chunk", r.URL.Path[10:])

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	if r.Method == "GET" {
		http.ServeFile(w, r, filepath.Join(path, "chunks", r.URL.Path[10:]))
	}
}

// uploadRepo logic.
func uploadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving repository")

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile(filepath.Join(path, "repository.knoxite"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored repository", filepath.Join(path, "repository.knoxite"))
}

// downloadRepo logic.
func downloadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving repository")

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(path, "repository.knoxite"))
}

func repository(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		downloadRepo(w, r)
	case "POST":
		uploadRepo(w, r)
	}
}

// uploadSnapshot logic.
func uploadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving snapshot")

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile(filepath.Join(path, "snapshots", handler.Filename), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored snapshot", filepath.Join(path, "snapshots", handler.Filename))
}

// downloadRepo logic.
func downloadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving snapshot", r.URL.Path[10:])

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(path, "snapshots", r.URL.Path[10:]))
}

func main() {
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/download/", download)
	http.HandleFunc("/repository", repository)
	http.HandleFunc("/snapshot", uploadSnapshot)
	http.HandleFunc("/snapshot/", downloadSnapshot)
	err := http.ListenAndServe(":42024", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
