package meta

import (
	mydb "fileupload/db"
)


type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}
// save to memory
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

// save to db
func UpdateFileMetaDB(fmeta FileMeta) bool {
  return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}


func GetFileMetaDB(fileSha1 string)( FileMeta,  error ){
	tFile, err := mydb.GetFileMeta(fileSha1)
	if tFile == nil || err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1: tFile.FileHash,
		FileName: tFile.FileName.String,
		FileSize: tFile.FileSize.Int64,
		Location: tFile.FileAddr.String,
	}
	return fmeta, nil
}


