package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			//Generate JSON for current state
			fmt.Fprint(w,"welcome")
		case "POST":
			// the FormFile function takes in the POST input id file
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
				fmt.Fprintf(w,html.EscapeString(name))
			}
		default:
			return
		}
	})
	s.ListenAndServe()
}
