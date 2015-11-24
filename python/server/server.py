#!/bin/python

import argparse
import flask
import hashlib
import io
import sys
import time

app = flask.Flask(__name__)
Files = {}
Content = {"Files": [],
           "Texts": [],
           "Info": {"SelfAddress": "",
                    "MaxUploadSize": "",
                    "ObjectTTL": ""}}
''' **************************************** 
    Utility 
    **************************************** '''


def setArguments(args):
    Content["Info"]["SelfAddress"] = args["SelfAddress"] + ":" + str(args["port"]) + "/"
    Content["Info"]["MaxUpoadSize"] = args["MaxUploadSize"]
    app.config['MAX_CONTENT_LENGTH'] = args["MaxUploadSize"]
    Content["Info"]["ObjectTTL"] = args["ObjectTTL"]


def getHash(string):
    toHash = (string + str(time.time())).encode('utf-8')
    md5hash = hashlib.md5(toHash).hexdigest()
    return (md5hash[:6])


def saveString(string):
    Content["Texts"].append({"Content": string,
                             "TimeCreated": int(time.time())})


def saveFile(uploaded_file):
    hashValue = getHash(uploaded_file.filename)
    # we have to store the components seperately, instead of leaving them in the
    # FileStorage object.
    # We do this because when we finish serving the request, the stream in the
    # FileStorage object is closed, and I do not know how to open it again.
    # Therefore I read the data and store it as an entry in the dictionary.
    # flask.send_file requires the data to have a "read" method, so for icing on
    # this hacky cake, I have to wrap the data in io.BytesIO.
    Files[hashValue] = {"Filename": uploaded_file.filename,
                        "Mimetype": uploaded_file.mimetype,
                        "Data": io.BytesIO(uploaded_file.read())}
    Content["Files"].append({"Name": uploaded_file.filename,
                             "TimeCreated": time.time(),
                             "Hash": hashValue,
                             "URL": (Content["Info"]["SelfAddress"] + hashValue
                                     )})


''' **************************************** 
    Routes
    **************************************** '''


@app.route('/', methods=['GET'])
def index():
    return (flask.jsonify(Content))


@app.route('/<hashValue>')
def download(hashValue):
    if hashValue in Files:
        file_data = Files[hashValue]
        return flask.send_file(file_data["Data"],
                               attachment_filename=file_data["Filename"],
                               mimetype=file_data["Mimetype"])
    else:
        flask.abort(404)


@app.route('/', methods=['POST'])
def upload():
    if flask.request.form:
        text = flask.request.form['text']
        if text:
            saveString(text)
    elif flask.request.files['file']:
        saveFile(flask.request.files['file'])
    return (flask.jsonify(Content))


''' **************************************** 
    Main
    **************************************** '''


def main(argv):
    parser = argparse.ArgumentParser(description='A python Uploader server')
    parser.add_argument('--port',
                        '-p',
                        action='store',
                        default='8080',
                        type=int)
    parser.add_argument('--MaxUploadSize',
                        '-us',
                        action='store',
                        default='2500000',
                        type=int)
    parser.add_argument('--ObjectTTL', action='store', default='300', type=int)
    parser.add_argument('--SelfAddress',
                        action='store',
                        default='http://localhost',
                        type=str)
    args = vars(parser.parse_args())
    setArguments(args)

    app.run(debug=True, port=args["port"])


if __name__ == '__main__':
    main(sys.argv[1:])
