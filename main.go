package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type SegyInfo struct {
	FileSize    int64
	SampleRate  int64
	TraceLength int64
	FormatCode  int64
	TotalTraces int64
	Token       string
}

func parseSegyFile(userFile string) (sampleRate int64, traceLength int64, formatCode int64, traceBytes int64, totalTraces int64) {
	fl, err := os.Open(userFile)
	defer fl.Close()
	if err != nil {
		fmt.Println(userFile, err)
		return
	}
	sampleRate, traceLength, formatCode, traceBytes, totalTraces = ParseSegyInfo(fl)
	return sampleRate, traceLength, formatCode, traceBytes, totalTraces
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method: ", r.Method)
	segyInfo := SegyInfo{}
	if r.Method == "GET" {
	} else if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		uf, handler, err := r.FormFile("uploadfile")
		if err == nil {
			defer uf.Close()

			fmt.Printf("%v\n", handler.Header)
			fileName := "data/" + handler.Filename
			tf, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer tf.Close()
			io.Copy(tf, uf)

			fileInfo, _ := tf.Stat()
			segyInfo.FileSize = fileInfo.Size()
			tf.Close()
			segyInfo.SampleRate, segyInfo.TraceLength, segyInfo.FormatCode, _, segyInfo.TotalTraces = parseSegyFile(fileName)
			fmt.Println("All things done!")
		}
	}

	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(crutime, 10))
	segyInfo.Token = fmt.Sprintf("%x", h.Sum(nil))

	{
		s1, _ := template.ParseFiles("doc/header.tmpl", "doc/content.tmpl", "doc/footer.tmpl")
		s1.ExecuteTemplate(w, "header", nil)
		s1.ExecuteTemplate(w, "content", segyInfo)
		s1.ExecuteTemplate(w, "footer", nil)
		s1.Execute(w, nil)
	}
}

func main() {
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
