/* This is JavaScript written by a C programmer, you have been warned. */
/* (Please help) */

var serverAddress = "http://localhost:8080/"

function UnixtoTwelveHour(timestamp) {
    var date = new Date(timestamp * 1000);
    var hours = date.getHours();
    var minutes = date.getMinutes();
    var ampm = hours >= 12 ? 'pm' : 'am';
    hours = hours % 12;
    hours = hours ? hours : 12;
    minutes = minutes < 10 ? '0' + minutes : minutes;
    return (hours + ':' + minutes + ' ' + ampm);
}

function ShowState(json) {
    console.log(json.Info.SelfAddress);
    console.log(json.Info.Location);
    console.log(json.Info.MaxUploadSize);
    console.log(json.Info.ObjectTTL);

    var ServerInfo = React.createClass({
        render: function() {
            return (React.createElement("div", null,
                React.createElement("h3", null, "Server Info"),
                React.createElement("p", null, "Server Address: ", json.Info.SelfAddress),
                React.createElement("p", null, "Server Location: ", json.Info.Location),
                React.createElement("p", null, "Max Server Upload Size (bytes): ",
                    json.Info.MaxUploadSize),
                React.createElement("p", null, "Time Objects will remain on server (seconds): ",
                    json.Info.ObjectTTL)));
        }
    });

    var TextList = React.createClass({
        render: function() {
            text_component = []
            if (json.Texts === null) {
                return (null)
            }
            $.each(json.Texts, function(index, element) {
                text_component.push(React.createElement("li", null, element.Content,
                    " ---  ", "Created: " + UnixtoTwelveHour(element.TimeCreated)))
            })
            return (React.createElement("div", null,
                React.createElement("h3", null, "Text List"),
                React.createElement("ul", null,
                    text_component)))
        }
    });

    var FileList = React.createClass({

        render: function() {
            file_component = []
            if (json.Files === null) {
                return (null)
            }
            $.each(json.Files, function(index, element) {
                file_component.push(React.createElement("li", null, element.Name,
                    " ---  ", React.createElement("a", {
                        href: element.URL
                    }, element.URL),
                    " --- ", "Created: " + UnixtoTwelveHour(element.TimeCreated)))
            })
            return (React.createElement("div", null,
                React.createElement("h3", null, "File List"),
                React.createElement("ul", null,
                    file_component)))
        }
    });

    var ServerContent = React.createClass({
        render: function() {
            return (
                React.createElement("div", null,
                    React.createElement("h3", null, "Server Content"),
                    React.createElement(TextList, null),
                    React.createElement(FileList, null)
                )
            );
        }
    });

    var Content = React.createClass({
        render: function() {
            return (React.createElement("div", null,
                React.createElement(ServerInfo, null, document.body),
                React.createElement(ServerContent, null, document.body)));
        }
    });


    var textForm = React.createClass({
        handleSubmit: function(e) {
            $("#textForm").submit(function(e) {
                var postData = $(this).serializeArray();
                var formURL = $(this).attr("action");
                $.ajax({
                    url: formURL,
                    type: "POST",
                    data: postData,
                    success: function(data, textStatus, jqXHR) {
                        ShowState(data)
                    },
                    error: function(jqXHR, textStatus, errorThrown) {
                        /* TODO We probably got here if the text message was empty
                         * or too large. Explain this more gracefully. */
                        alert("fail")
                    }
                });
                e.preventDefault()
            });
            $("#textForm").submit()
            e.preventDefault()
        },
        render: function() {
            return (React.createElement("form", {
                    id: "textForm",
                    //TODO let user set this
                    action: serverAddress,
                    method: "post",
                    encType: "multipart/form-data",
                    onSubmit: this.handleSubmit
                },
                React.createElement("label", {
                    htmlFor: "text"
                }, "Say Something:"),
                React.createElement("input", {
                    type: "text",
                    name: "text",
                    id: "text",
                }),
                React.createElement("button", null, {
                    value: "submit"
                })));
        }
    });
    var fileForm = React.createClass({
        handleSubmit: function(e) {
            $("#fileForm").submit(function(e) {
                var postData = $(this).serializeArray();
                var formURL = $(this).attr("action");
                $.ajax({
                    url: formURL,
                    type: "POST",
                    data: new FormData(this),
                    success: function(data, textStatus, jqXHR) {
                        ShowState(data)
                    },
                    error: function(jqXHR, textStatus, errorThrown) {
                        /* TODO We probably got here because the file being
                         * uploaded was too big. Explain this more
                         * gracefully. */
                        alert(errorThrown)
                    },
                    processData: false,
                    contentType: false
                });
                e.preventDefault()
            });
            $("#fileForm").submit()
            e.preventDefault()
        },
        render: function() {
            return (React.createElement("form", {
                    id: "fileForm",
                    //TODO let user set this
                    action: serverAddress,
                    method: "post",
                    encType: "multipart/form-data",
                    onSubmit: this.handleSubmit
                },
                React.createElement("label", {
                    htmlFor: "file"
                }, "Filename:"),
                React.createElement("input", {
                    type: "file",
                    name: "file",
                    id: "file"
                }),
                React.createElement("button", null, {
                    value: "submit"
                })));
        }
    });

    var Forms = React.createClass({
        render: function() {
            return (React.createElement("div", null,
                React.createElement(textForm, null, document.body),
                React.createElement(fileForm, null, document.body)));
        }
    });

    var Server = React.createClass({
        render: function() {
            return (React.createElement("div", null,
                React.createElement(Content, null, document.body),
                React.createElement(Forms, null, document.body)));
        }
    });

    React.render(React.createElement(Server, null), document.body);
}

$.getJSON(serverAddress, ShowState).fail(function() {
    //TODO Improve display of this error
    alert("Server is not running?");
})
