package handler

import (
	"encoding/json"
	"fileupload/meta"
	"fileupload/util"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)



func UploadHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET" {
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		file, head, err := r.FormFile("file")
		if err != nil {
			log.Printf("Failed to get data, err: %s\n", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location:"./tmp/"+head.Filename,
			UploadAt: time.Now().Format("2006-01-06 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			log.Printf("ile to create file, err: %s\n", err.Error())
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			log.Printf("Failed to save data into file, err: %s\n", err.Error())
			return
		}
		//  newFile 文件对象的读写位置重置到文件的开头，这样你可以从文件的开头再次读取数据，而不是在文件的当前位置继续读取。这
	  newFile.Seek(0,0)

	  fileMeta.FileSha1 = util.FileSha1(newFile)
		log.Printf("file name %s : %s",fileMeta.FileName, fileMeta.FileSha1)
		meta.UpdateFileMeta(fileMeta)
   http.Redirect(w, r, "./file/upload/suc", http.StatusFound)
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

func GetFileMetaHandler( w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	 filehash := r.URL.Query().Get("filehash")
   fMeta := meta.GetFileMeta(filehash)
	 data, err := json.Marshal(fMeta)
	 if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	 }
	 w.Write(data)
}