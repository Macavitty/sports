package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const UplPrefix = "upl-"
const noPermsError = "[-] Unable to create the file for writing. Check your write access privilege."

func uploadVideo(w http.ResponseWriter, r *http.Request) {
	//if r.URL.Path != "/" {
	//	http.NotFound(w, r)
	//	return
	//}
	//contentType := r.Header.Get("Content-type")


	file, header, err := r.FormFile("file")

	if err != nil {
		log.Println("[-] Error in r.FormFile ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{'error': %s}", err)
		return
	}
	defer file.Close()

	out, err := os.Create("uploaded/" + UplPrefix + header.Filename)
	if err != nil {
		log.Println(noPermsError, err)
		_, _ = fmt.Fprintf(w, "%s" + noPermsError, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer out.Close()

	// save from POST
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("[-] Error saving file.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("[+] File uploaded successfully: " +UplPrefix, header.Filename)
}


func run(){
	http.HandleFunc("/", uploadVideo)
	http.ListenAndServe(":15000", nil)
	//go http.ListenAndServe(":15000", nil)

}


func main() {
	_ = os.Mkdir("uploaded", os.ModePerm)
	run()
}
