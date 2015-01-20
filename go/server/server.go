package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type JsonObject struct {
	Files []File
	Texts []Text
	Info  ServerInfoObject
}

type File struct {
	Name        string
	TimeCreated int64
	Hash        string
	URL         string
	ContentType string
}

type Text struct {
	Content     string
	TimeCreated int64
}

type ServerInfoObject struct {
	SelfAddress   string
	MaxUploadSize int
	ObjectTTL     int64
}

var Contents JsonObject

var MaxUploadSize string

var FilesStorage = make(map[string]*bytes.Buffer)

func OutOfDate(timeCreated int64) bool {
	CurrentTime := time.Now().Unix()
	return CurrentTime-timeCreated >= Contents.Info.ObjectTTL
}

/* Iterate over Contents and remove items older than TTL */
func RemoveExpiredItems() {
	var UnexpiredFiles []File
	var UnexpiredTexts []Text
	for _, file := range Contents.Files {
		if OutOfDate(file.TimeCreated) {
			delete(FilesStorage, file.Hash)
		} else {
			UnexpiredFiles = append(UnexpiredFiles, file)
		}
	}
	Contents.Files = UnexpiredFiles

	for _, text := range Contents.Texts {
		if !OutOfDate(text.TimeCreated) {
			UnexpiredTexts = append(UnexpiredTexts, text)
		}
	}
	Contents.Texts = UnexpiredTexts
}

func InitContents(selfAddr string, MaxUploadSize int, TTL int64) {
	Contents = JsonObject{
		Info: ServerInfoObject{
			SelfAddress:   selfAddr,
			MaxUploadSize: MaxUploadSize,
			ObjectTTL:     TTL,
		},
	}
	_, err := json.Marshal(Contents)
	if err != nil {
		fmt.Println("Failed during initialization")
		os.Exit(1)
	}
}

func GenRandomString(FileName string) string {
	NumChars := 6
	hash := md5.New()
	io.WriteString(hash, strconv.FormatInt(time.Now().UnixNano(), 10)+FileName)
	return fmt.Sprintf("%x", hash.Sum(nil))[:NumChars]
}

func MainResponse(w http.ResponseWriter, r *http.Request) {
	if size, _ := strconv.Atoi(r.Header.Get("Content-Length")); size > Contents.Info.MaxUploadSize {
		http.Error(w, "File too large", 413)
		return
	}
	/* Set header to enable AJAX for clients */
	w.Header().Set("Access-Control-Allow-Origin", "*")
	RemoveExpiredItems()
	switch r.Method {
	case "GET":
		InputUrl := r.URL.String()[1:]
		if InputUrl == "" {
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			//TODO Save file with actuall name instead of hash
			if err == nil {
				w.Write(obj)
			} else {
				http.Error(w, "Coudln't marshall json", 500)
				os.Exit(1)
			}
		} else if FilesStorage[InputUrl] == nil {
			http.NotFound(w, r)
		} else {
			io.Copy(w, FilesStorage[InputUrl])
		}
		return
	case "POST":
		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			NewRandomString := GenRandomString(header.Filename)
			b := &bytes.Buffer{}
			_, err = io.Copy(b, file)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			FilesStorage[NewRandomString] = b
			Contents.Files = append(Contents.Files, File{
				Name:        html.EscapeString(header.Filename),
				TimeCreated: time.Now().Unix(),
				Hash:        NewRandomString,
				URL:         Contents.Info.SelfAddress + NewRandomString,
				ContentType: header.Header.Get("Content-Type"),
			})
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			if err == nil {
				w.Write(obj)
			} else {
				fmt.Fprintln(w, err)
			}

		} else {
			content := r.FormValue("text")
			if content == "" {
				http.Error(w, "Input text cannot be empty", 400)
				return
			}
			Contents.Texts = append(Contents.Texts, Text{
				Content:     content,
				TimeCreated: time.Now().Unix(),
			})
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			if err == nil {
				w.Write(obj)
			} else {
				http.Error(w, "Couldn't marshal json for text", 500)
				os.Exit(1)
			}
		}
	default:
		return
	}
}

func main() {
	DefaultPort := "8080"
	DefaultAddr := "http://localhost:" + DefaultPort + "/"
	DefaultTTL := int64(300)
	DefaultUploadSize := 2500000

	PortPtr := flag.String("port", DefaultPort, "Port to run server on")
	SelfAddrPtr := flag.String("selfAddr", DefaultAddr, "URL Address to access server")
	TTLPtr := flag.Int64("TTL", DefaultTTL, "Time files and texts stay on server")
	MaxUploadSizePtr := flag.Int("MaxUploadSize", DefaultUploadSize, "Maximum size of file uploaded")
	flag.Parse()
	s := &http.Server{
		Addr:           ":" + *PortPtr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	InitContents(*SelfAddrPtr, *MaxUploadSizePtr, *TTLPtr)
	http.HandleFunc("/", MainResponse)
	panic(s.ListenAndServe())
}
