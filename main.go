package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	//pb "transfer.proto"
	//"golang.org/x/net/context"
	//"google.golang.org/grpc"
)

const UplPrefix = "upl-"
const noPermsError = "[-] Unable to create the file for writing. Check your write access privilege."

func getResult() string {
	// pass video to python

	/*
		adress := "localhost:66666"
		connection, err := grpc.Dial(adress, grpc.WithInsecure())
		if err != nil{
			log.Fatal("Error: can`t connect to python server")
		}
		defer connection.Close()
		// create grpc client
		client := pb.NewNetRequestClient(connection)
		message := &pb.NetRequest{Query: }
		response, err := client.Func()
	*/

	var percentFromPython = 42.73
	var percent = strconv.FormatFloat(percentFromPython, 'f', 2, 64)
	return percent
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		returnVideo(w, r)
	} else {
		uploadVideo(w, r)
	}
}

/*
* BEWARE
* for some reason curl (curl http://localhost:15000/name.mp4 -o name.mp4)
* is not able to download video properly (Truncated file - missing n bytes)
* please use wget instead: wget http://localhost:15000/name.mp4
*
* you also can check what`s wrong with the video with MP4Box:
* MP4Box -info -v path/to/video_from_server
 */
func returnVideo(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	videoName := r.URL.String()[1:]
	realVideoName := "uploaded/" + UplPrefix + videoName

	log.Println("Client requests: " + realVideoName)

	// Check if file exists and open
	video, err := os.Open(realVideoName)

	if err != nil {
		//File not found, send 404
		log.Println("[-] Error in GET: required video not exists")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer video.Close() //Close after function return

	// create and send the correct headers

	// Get the Content-Type of the file
	// Create a buffer to store the header of the file in
	FileHeader := make([]byte, 64)
	// Copy the headers into the FileHeader buffer
	_, err = video.Read(FileHeader)
	if err != nil {
		log.Println("[-] Error: could not open video")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//content type of file
	//FileContentType := http.DetectContentType(FileHeader)

	//file size
	FileStat, _ := video.Stat()                        // info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; video name="+videoName)
	w.Header().Set("Content-Type", "multipart/form-data")
	w.Header().Set("Content-Length", FileSize)

	//read 512 bytes from the file already, so we reset the offset back to 0
	_, _ = video.Seek(0, 0)
	_, err = io.Copy(w, video)
	if err != nil {
		log.Println("Error in GET: could not put video to responseWriter")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("[+] GET successful: " + videoName)
	return

}

/*
* use curl to test:
* curl -F "file=@/path/to/video" http://localhost:15000/
 */
func uploadVideo(w http.ResponseWriter, r *http.Request) {

	clientVideo, header, err := r.FormFile("file")

	if err != nil {
		log.Println("[-] Error in r.FormFile ", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "{'error': %s}", err)
		return
	}
	defer clientVideo.Close()

	out, err := os.Create("uploaded/" + UplPrefix + header.Filename)
	if err != nil {
		log.Println(noPermsError, err)
		_, _ = fmt.Fprintf(w, "%s"+noPermsError, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer out.Close()

	// save from POST
	_, err = io.Copy(out, clientVideo)
	if err != nil {
		log.Println("[-] Error saving clientVideo.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("[+] File uploaded successfully: " + UplPrefix + header.Filename)

	// return user result
	w.Header().Set("UserResult", getResult())

}

func run() {
	http.HandleFunc("/", requestHandler)
	_ = http.ListenAndServe(":15000", nil)
	//go http.ListenAndServe(":15000", nil)

}

func main() {
	log.Println("* Server running on port 15000 *")
	_ = os.Mkdir("uploaded", os.ModePerm)
	run()
}
