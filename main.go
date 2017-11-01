package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const html = `

<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset=utf-8>
    <meta name="viewport" content="width=device-width, initial-scale=0.41, maximum-scale=1" />
    <title>Simple Upload</title>
    <style type="text/css">
    * {
        color: white;
        font-family: sans-serif;
        padding: 0;
        margin: 0;
        cursor: pointer;
        -webkit-touch-callout: none;
        -webkit-user-select: none;
        -khtml-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
        user-select: none;
    }

    #upload {
        width: 100%;
        top: 0;
        position: absolute;
        background-color: #f44242;
        height: 100%;
    }

    .text {
        position: absolute;
        top: 40%;
        text-align: center;
        width: 100%;
        font-size: 3em;
    }

    input[type="file"] {
        display: none;
    }
    </style>
</head>

<body>
    <!-- <input type='file' id='file' />   -->
    <form id="file-form" enctype="multipart/form-data" action="/upload" method="post">
        <label>
            <div id="upload">
                <div class="text" id="uploadtext">UPLOAD<br>({{ . }}:8080)</div>
            </div>
            <input type="file" id='file' name="uploadfile" />
        </label>
        <input type="submit" value="upload" />
    </form>
    <script>
    var uploadtext = document.getElementById('uploadtext');
    var fileelt = document.getElementById('file');

    document.body.ondragover = function() { uploadtext.innerHTML = 'DROP YOUR FILE HERE'; return false; };
    document.body.ondrop = function(e) {
        e.preventDefault();
        readfiles(e.dataTransfer.files);
    };
    fileelt.addEventListener("change", function(e) { 
        var formData = new FormData(document.getElementById("file-form"));
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/upload');
        xhr.onload = function() { uploadtext.innerHTML =  xhr.responseText; };
        xhr.upload.onprogress = function(event) {
            if (event.lengthComputable) {
                var complete = (event.loaded / event.total * 100 | 0);
                var complete36 = (event.loaded / event.total * 36 | 0);
                uploadtext.innerHTML = 'UPLOADING<br>PROGRESS ' + complete + '%';
            }
        };
        xhr.send(formData);
    });
    </script>
</body>

</html>

`

var ip = ""

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.New("upload").Parse(html)
		t.Execute(w, ip)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "ERROR")
			return
		}
		defer f.Close()
		io.Copy(f, file)
		fmt.Fprintf(w, "UPLOAD COMPLETE!")
	}
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}
		}
	}

	go open("http://" + ip + ":8080")

	http.HandleFunc("/", upload)
	err = http.ListenAndServe(":8080", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
