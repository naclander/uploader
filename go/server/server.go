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
	Url         string
}

type Text struct {
	Content     string
	TimeCreated int64
}

type ServerInfoObject struct {
	SelfAddress string
	Location    string
	ObjectTTL   int64
}

var Contents JsonObject

var FilesStorage = make(map[string]*bytes.Buffer)

/* Iterate over Contents and remove items older than TTL */
func RemoveExpiredItems() {
	for i := 0; i < len(Contents.Files); i++ {
		if time.Now().Unix()-Contents.Files[i].TimeCreated >
			Contents.Info.ObjectTTL {
			delete(FilesStorage, Contents.Files[i].Hash)
			Contents.Files = append(Contents.Files[:i], Contents.Files[i+1:]...)
		}
	}
	for i := 0; i < len(Contents.Texts); i++ {
		if time.Now().Unix()-Contents.Texts[i].TimeCreated >
			Contents.Info.ObjectTTL {
			Contents.Texts = append(Contents.Texts[:i], Contents.Texts[i+1:]...)
		}
	}
}

func InitContents(location, selfAddr string, TTL int64) {
	Contents = JsonObject{
		Info: ServerInfoObject{
			SelfAddress: selfAddr,
			Location:    location,
			ObjectTTL:   TTL,
		},
	}
	obj, err := json.Marshal(Contents)
	//TODO Handle error
	if err == nil {
		os.Stdout.Write(obj)
	}
}

func GenRandomString(FileName string) string {
	NumChars := 6
	hash := md5.New()
	io.WriteString(hash, strconv.FormatInt(time.Now().UnixNano(), 10)+FileName)
	return fmt.Sprintf("%x", hash.Sum(nil))[:NumChars]
}

func MainResponse(w http.ResponseWriter, r *http.Request) {
	RemoveExpiredItems()
	switch r.Method {
	case "GET":
		InputUrl := r.URL.String()[1:]
		if InputUrl == "" {
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			//TODO Handle error
			if err == nil {
				w.Write(obj)
			}
		} else if FilesStorage[InputUrl] == nil {
			http.NotFound(w, r)
			return
		} else {
			fmt.Fprintf(w, "It exists!!!")
		}
	case "POST":
		file, header, err := r.FormFile("file")
		fmt.Println("Read file and header")

		if err == nil {
			fmt.Println("Uploading binary file")
			defer file.Close()

			out, err := os.Create("/tmp/uploadedfile")
			if err != nil {
				fmt.Fprintf(w, "Unable to create the file for writing")
				return
			}

			defer out.Close()

			// write the content from POST to the file

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
				Url:         Contents.Info.SelfAddress + NewRandomString,
			})
			fmt.Println(Contents)
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			os.Stdout.Write(obj)
			if err == nil {
				w.Write(obj)
			} else {
				fmt.Fprintln(w, err)
			}

		} else {
			name := r.FormValue("text")
			Contents.Texts = append(Contents.Texts, Text{
				Content:     name,
				TimeCreated: time.Now().Unix(),
			})
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			//TODO Handle error
			if err == nil {
				w.Write(obj)
			}
			/*
				obj,err := json.Marshal(Contents)
				if err == nil{
					os.Stdout.Write(obj)
				}
			*/
		}
	default:
		return
	}
}

func main() {
	PortPtr := flag.String("port", "8080", "Port to run server on")
	LocPtr := flag.String("location", "NA", "Server Geographical Location")
	SelfAddrPtr := flag.String("selfAddr", "localhost:"+*PortPtr+"/",
		"URL Address to access server")
	TTLPtr := flag.Int64("TTL", 300, "Time files and texts stay on server")
	flag.Parse()
	s := &http.Server{
		Addr:           ":" + *PortPtr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	InitContents(*LocPtr, *SelfAddrPtr, *TTLPtr)
	http.HandleFunc("/", MainResponse)
	panic(s.ListenAndServe())
}
