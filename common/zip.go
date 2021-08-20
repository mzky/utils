package common

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/mzky/zip"
)

func IsZip(zipPath string) bool {
	f, err := os.Open(zipPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// Zip password值可以为空""
func Zip(zipPath, password string, fileList []string) error {
	fz, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer fz.Close()
	zw := zip.NewWriter(fz)
	defer zw.Close()

	for _, fileName := range fileList {
		fr, errA := os.Open(fileName)
		if errA != nil {
			return errA
		}

		// 写入文件的头信息
		var w io.Writer
		var errB error
		if password != "" {
			w, errB = zw.Encrypt(fileName, password, zip.AES256Encryption)
		} else {
			w, errB = zw.Create(fileName)
		}

		if errB != nil {
			return errB
		}

		// 写入文件内容
		_, errC := io.Copy(w, fr)
		if errC != nil {
			return errC
		}
		_ = fr.Close()
	}
	return zw.Flush()
}

// UnZip password值可以为空""
// 当decompressPath值为"./"时，解压到相对路径
func UnZip(zipPath, password, decompressPath string) error {
	if !FileIsExist(zipPath) {
		return errors.New("找不到压缩文件")
	}
	if !IsZip(zipPath) {
		return errors.New("压缩文件格式不正确或已损坏")
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if password != "" {
			if f.IsEncrypted() {
				f.SetPassword(password)
			} else {
				return errors.New("must be encrypted")
			}
		}

		fp := filepath.Join(decompressPath, f.Name)
		_ = os.MkdirAll(filepath.Dir(fp), os.ModePerm)

		w, errA := os.Create(fp)
		if errA != nil {
			return errors.New("无法创建解压文件")
		}
		fr, errB := f.Open()
		if errB != nil {
			return errors.New("解压密码不正确")
		}
		if _, errC := io.Copy(w, fr); errC != nil {
			return errC
		}
		_ = fr.Close()
		_ = w.Close()
	}
	return nil
}
