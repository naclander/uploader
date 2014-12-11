package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

type JsonObject struct {
	Files []File
	Texts []Text
	Info  ServerInfoObject
}

type File struct {
	name string
	TimeCreated int64
	url  string
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

/* Iterate over Contents and remove items older than TTL */
func RemoveExpiredItems() {
	for i := 0; i < len(Contents.Files); i++ {
		if (time.Now().Unix() - Contents.Files[i].TimeCreated > 
		   Contents.Info.ObjectTTL) {
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
			RemoveExpiredItems()
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
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	InitContents("192.168.LOL.LOL", "USAUSAUSA", 60)
	http.HandleFunc("/", MainResponse)
	s.ListenAndServe()
}
