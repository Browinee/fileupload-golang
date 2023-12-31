package main

import (
	"fileupload/handler"
	"fileupload/middleware"
	"fmt"
	"net/http"
)


func main(){

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/query", middleware.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)
	http.HandleFunc("/file/fastupload", handler.TryFastUploadHandler)

	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SigninHandler)
	http.HandleFunc("/user/info", middleware.HTTPInterceptor(handler.UserInfoHandler))

	// NOTE: multi-part
	http.HandleFunc("/file/mpupload/init", middleware.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart",middleware.HTTPInterceptor( handler.UploadPartHandler ))
	http.HandleFunc("/file/mpupload/complete", middleware.HTTPInterceptor( handler.CompleteUploadHandler ))


	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Fail to start server, err %s", err.Error())
	}
}