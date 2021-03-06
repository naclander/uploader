package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
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

func die(message string) {
	log.Fatal(message)
	os.Exit(1)
}

func isExpired(timeCreated int64) bool {
	return (time.Now().Unix()-timeCreated >= Contents.Info.ObjectTTL)
}

/* Remove objects alive longer than TTL */
func RemoveExpiredItems() {
	var UnexpiredFiles []File
	var UnexpiredTexts []Text
	for _, file := range Contents.Files {
		if isExpired(file.TimeCreated) {
			delete(FilesStorage, file.Hash)
		} else {
			UnexpiredFiles = append(UnexpiredFiles, file)
		}
	}
	Contents.Files = UnexpiredFiles
	for _, text := range Contents.Texts {
		if !isExpired(text.TimeCreated) {
			UnexpiredTexts = append(UnexpiredTexts, text)
		}
	}
	Contents.Texts = UnexpiredTexts
}

func InitContents(selfAddr, port string, MaxUploadSize int, TTL int64) {
	Contents = JsonObject{
		Info: ServerInfoObject{
			SelfAddress:   selfAddr + ":" + port + "/",
			MaxUploadSize: MaxUploadSize,
			ObjectTTL:     TTL,
		},
	}
	_, err := json.Marshal(Contents)
	if err != nil {
		die("Failed during initialization")
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
		requestURL := r.URL.String()[1:]
		/* Client requested entire json object */
		if requestURL == "" {
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			if err == nil {
				w.Write(obj)
			} else {
				message := "Couldn't marshall json"
				http.Error(w, message, 500)
				die(message)
			}
			/* Client requested a file that no longer exists or never existed */
		} else if FilesStorage[requestURL] == nil {
			http.NotFound(w, r)
			/* Client requested a file we have */
		} else {
			retrievedFile := (*FilesStorage[requestURL])
			/* Need to create a Reader so we can send the file again next time
			 * It is requested. */
			written, err := io.Copy(w, bytes.NewReader(retrievedFile.Bytes()))
			if written != int64(retrievedFile.Len()) || err != nil {
				die("Couldn't send back entire file")
			}
		}
	case "POST":
		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			NewRandomString := GenRandomString(header.Filename)
			b := &bytes.Buffer{}
			maxUploadSize := int64(Contents.Info.MaxUploadSize)
			written, err := io.CopyN(b, file, maxUploadSize)
			if err != io.EOF && written != maxUploadSize {
				message := "Couldn't read file"
				http.Error(w, message, 500)
				die(message)
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
				message := "Couldn't marshal json for text"
				http.Error(w, message, 500)
				die(message)
			}
		}
	default:
		return
	}
}

func main() {
	DefaultPort := "8080"
	DefaultAddr := "http://localhost"
	DefaultTTL := int64(300)     /* 5 Minutes */
	DefaultUploadSize := 2500000 /* 2.5 Megabytes */

	PortPtr := flag.String("port", DefaultPort, "Port to run server on")
	SelfAddrPtr := flag.String("selfAddr", DefaultAddr, "URL Address to access server")
	TTLPtr := flag.Int64("TTL", DefaultTTL, "Time files and texts stay on server")
	MaxUploadSizePtr := flag.Int("MaxUploadSize", DefaultUploadSize,
		"Maximum size of file uploaded")
	flag.Parse()
	s := &http.Server{
		Addr:           ":" + *PortPtr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//TOOD  Limit the upload size explicitly
	InitContents(*SelfAddrPtr, *PortPtr, *MaxUploadSizePtr, *TTLPtr)
	http.HandleFunc("/", MainResponse)
	panic(s.ListenAndServe())
}
