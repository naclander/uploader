# Uploader

![](https://imgs.xkcd.com/comics/file_transfer.png)

## About
Its an easy way to share notes and files between different people.

Uploader lets you quickly write sentences, or upload files for others to quickly
look at or download. All sentences and files uploaded are temporary, and get
automatically deleted after a specified amount of time.

<!---
TODO:
Provide a link here to demo, once there is a stable version of the application.
-->

## General Idea
Uploader is split between a client component, and a server component. These
components should work independently from each other, and make no assumptions
about the inner workings of the component they are interacting with, other
than the established API.

One should be be able to mix and match any client and server implementations.

## API

## Client
The client should be able to upload files and text strings to the server.

* files - POST using `type=file`
* text string - POST using `type=text`

The server will respond with a JSON object showing the current state of all
of the files it contains (Name, URL, timestamp, etc.), the format is discussed
later.

It is up to the client on how this JSON format is interpreted and shown to the
user. For example, a web client might use the JSON object to display HTML
after every POST, where as a CLI client might not. This JSON object is also
returned when a GET is made to the base server address.

## Server
The server should be able to accept files and text strings uploaded to it
by the client, see the Client API section for more details.

The server should keep track of the text strings and files uploaded to it, and
return a JSON object after every POST or GET from the client. The server should
be configured with a TTL for each file uploaded. The server should delete items
that exist longer than the TTL.

The server should generate a URL for any file that has been uploaded to it. This
URL is included in the JSON object, and when a GET is made to this URL, the server
should send over the file corresponding to that URL.

Some more server configuration information:

* TTL - how long files and text strings will remain on the server
* MAX Upload Size - maximum file size allowed by the server

## Example usage

Example usage of an Uploader CLI client:

```
>> uploader
No files to show
No texts to show
>> uploader -t "This is a test string"
>> uploader -f ~/.vimrc
File URL:
http://serverurl:8000/asf3da
>> uploader
Files:
10:07:43 .vimrc http://serverurl:8000/asf3da
Texts:
10:07:32 This is a text string
```

### JSON Object Format
Example of a JSON object sent from the server to the client:

```
{
   "Files":[
      {
         "Name":"document1.pdf",
         "TimeCreated":1418341620,
         "Hash":"b8b53e",
         "URL":"http://example.com/b8b53e"
      },
      {
         "Name":"document(2).pdf",
         "TimeCreated":1418341622,
         "Hash":"d28586",
         "URL":"http://example.com/d28586"
      }
   ],
   "Texts":[
      {
         "Content":"String 1",
         "TimeCreated":1418341616
      },
      {
         "Content":"String 2",
         "TimeCreated":1418341618
      }
   ],
   "Info":{
      "SelfAddress":"192.168.2.1",
      "MaxUploadSize":"2048",
      "ObjectTTL":300
   }
}
```
