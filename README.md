# Uploader

##About
Uploader lets you quickly write sentences, or upload files for others to quickly
look at or download. All sentences and files uploaded are temporary, and get
automatically deleted after a specified amount of time.

Its an easy way to share notes and files between different people.

<!---
TODO:
Provide a link here to demo, once there is a stable version of the application.
-->

##General Idea
Uploader is split between a client component, and a server component. These
components should work independently from each other, and make no assumptions
about the inner workings of the component they are interacting with, other
than the established API.

One should ideally be able to run any combination of client and server
components.

###Client API
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

###Server API
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
* MAX Size - maximum file size allowed by the serverjjj

###JSON Object Format
Below is an example of a JSON object sent from the server to the client:

```
{
   "Files":[
      {
         "Name":"document1.pdf",
         "TimeCreated":1418341620,
         "Hash":"b8b53e",
         "Url":"http://example.com/b8b53e"
      },
      {
         "Name":"document(2).pdf",
         "TimeCreated":1418341622,
         "Hash":"d28586",
         "Url":"http://example.com/d28586"
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
      "IPAdress":"192.168.2.1",
      "Location":"USA",
      "ObjectTTL":300
   }
}
```
