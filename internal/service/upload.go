package service

import (
	"mime/multipart"
	"os"

	"github.com/google/uuid"
	"github.com/gookit/goutil/dump"
	"github.com/pkg/errors"

	"github.com/haierspi/golang-image-upload-service/global"
	"github.com/haierspi/golang-image-upload-service/pkg/oss"
	"github.com/haierspi/golang-image-upload-service/pkg/upload"
)

type FileInfo struct {
	ImageTitle string `json:"imageTitle"`
	ImageUrl   string `json:"imageUrl"`
}

func (svc *Service) UploadFile(fileType upload.FileType, file multipart.File, fileHeader *multipart.FileHeader) (*FileInfo, error) {

	var accessUrlPre string
	var fileName string

	if fileHeader.Filename == "image.png" {
		fileName = upload.GetFileName(uuid.New().String() + fileHeader.Filename)
	} else {
		fileName = upload.GetFileName(fileHeader.Filename)
	}

	dump.P(fileHeader.Filename)

	if !upload.CheckContainExt(fileType, fileName) {
		return nil, errors.New("file suffix is not supported.")
	}
	if upload.CheckMaxSize(fileType, file) {
		return nil, errors.New("exceeded maximum file limit.")
	}

	uploadSavePath := upload.GetSavePath()
	if upload.CheckSavePath(uploadSavePath) {
		if err := upload.CreateSavePath(uploadSavePath, os.ModePerm); err != nil {
			return nil, errors.New("failed to create save directory.")
		}
	}
	if upload.CheckPermission(uploadSavePath) {
		return nil, errors.New("insufficient file permissions.")
	}

	dateDirFileName := upload.GetSavePreDirPath() + fileName
	if err := upload.SaveFile(fileHeader, uploadSavePath+"/"+dateDirFileName); err != nil {
		return nil, err
	}
	accessUrlPre = global.AppSetting.UploadServerUrl

	// 阿里云oss
	if global.OSSSetting.Enable {
		err := oss.UploadByFile(dateDirFileName, file)
		if err != nil {
			return nil, errors.Wrap(err, "oss.UploadByFile err")
		}
	}

	accessUrl := accessUrlPre + "/" + dateDirFileName

	return &FileInfo{ImageTitle: "", ImageUrl: accessUrl}, nil
}
