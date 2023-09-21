package handler

import (
	"encoding/json"
	db "fileupload/db"
	"fileupload/meta"
	"fileupload/util"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
		//  newFile 文件对象的读写位置重置到文件的开头，这样你可以从文件的开头再次读取数据，而不是在文件的当前位置继续读取。
	  newFile.Seek(0,0)
	  fileMeta.FileSha1 = util.FileSha1(newFile)
		// NOTE: save to memory
		// meta.UpdateFileMeta(fileMeta)
		// NOTE: save to db
   meta.UpdateFileMetaDB(fileMeta)
	 // NOTE: update user file table
	 r.ParseForm()
	 username := r.Form.Get("username")
	 suc := db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	 if suc {
		 http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	 } else {
		 w.Write([]byte("Upload Failed!"))
	 }
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

func GetFileMetaHandler( w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	 filehash := r.URL.Query().Get("filehash")
  //  fMeta := meta.GetFileMeta(filehash)
   fMeta, err := meta.GetFileMetaDB(filehash)
	 if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	 }

	 data, err := json.Marshal(fMeta)
	 if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	 }
	 w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)
	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)

}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "Fail to fast upload. Please use normal upload",
		}
		w.Write(resp.JSONBytes())
		return
	}

	suc := db.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "Fast upload successfully.",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "Fail to fast upload. Please try later.",
	}
	w.Write(resp.JSONBytes())
	return

}