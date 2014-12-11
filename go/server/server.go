package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"
	"flag"
)

type JsonObject struct {
	Files []File
	Texts []Text
	Info  ServerInfoObject
}

type File struct {
	Name string
	TimeCreated int64
	Hash string
	Url  string
}

type Text struct {
	Content string
	TimeCreated int64
}

type ServerInfoObject struct {
	IPAdress string
	Location string
	ObjectTTL int64
}

var Contents JsonObject

var FilesStorage = make(map[string]*bytes.Buffer)

/* Iterate over Contents and remove items older than TTL */
func RemoveExpiredItems() {
	for i := 0; i < len(Contents.Files); i++ {
		if (time.Now().Unix() - Contents.Files[i].TimeCreated > 
		   Contents.Info.ObjectTTL) {
			delete(FilesStorage, Contents.Files[i].Hash)
			Contents.Files = append(Contents.Files[:i], Contents.Files[i+1:]...)
		}	
	}	
	for i := 0; i < len(Contents.Texts); i++ {
		if (time.Now().Unix() - Contents.Texts[i].TimeCreated > 
		   Contents.Info.ObjectTTL) {
			Contents.Texts = append(Contents.Texts[:i], Contents.Texts[i+1:]...)
		}	
	}	
}

func InitContents(addr, location string, TTL int64) {
	Contents = JsonObject{
		Info: ServerInfoObject{
			IPAdress: addr,
			Location: location,
			ObjectTTL: TTL,
		},
	}
	obj, err := json.Marshal(Contents)
	//TODO Handle error
	if err == nil {
		os.Stdout.Write(obj)
	}
}

func MainResponse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		RemoveExpiredItems()
		w.Header().Set("Content-Type", "application/json")
		obj, err := json.Marshal(Contents)
		//TODO Handle error
		if err == nil{
			w.Write(obj)
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

			NewRandomString := GenRandomString()
			fmt.Printf("New string is %s\n\n\n",NewRandomString)
			b := &bytes.Buffer{}
			_, err = io.Copy(b,file)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			FilesStorage[NewRandomString] = b
			NewFile := File{
				Name: html.EscapeString(header.Filename),
				TimeCreated: time.Now().Unix(),
				Hash: NewRandomString,
				Url: "lol",
			}
			Contents.Files = append(Contents.Files, NewFile)
			fmt.Println(Contents)
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			os.Stdout.Write(obj)

			if err == nil{
				w.Write(obj)
			} else{
				fmt.Fprintln(w, err)
			}

			_, err = io.Copy(out, file)

			if err != nil {
				fmt.Fprintln(w, err)
			}

			fmt.Fprintf(w, "File uploaded successfully : ")
			fmt.Fprintf(w, html.EscapeString(header.Filename))

		} else {
			name := r.FormValue("text")
			NewText := Text{
				Content: name,
				TimeCreated: time.Now().Unix(),
			}
			Contents.Texts = append(Contents.Texts,NewText)
			w.Header().Set("Content-Type", "application/json")
			obj, err := json.Marshal(Contents)
			//TODO Handle error
			if err == nil{
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
	AddrPtr := flag.String("address", "NA", "Server IP Address")
	PortPtr := flag.String("port", "8080", "Port to run server on")
	LocPtr := flag.String("location", "NA", "Server Geographical Location")
	TTLPtr := flag.Int64("TTL", 300, "Time files and texts stay on server")
	flag.Parse()
	s := &http.Server{
		Addr:           ":" + *PortPtr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	InitContents(*AddrPtr, *LocPtr, *TTLPtr)
	http.HandleFunc("/", MainResponse)
	panic(s.ListenAndServe())
}
