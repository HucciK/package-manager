package internal

import (
	"archive/zip"
	"fmt"
	"github.com/HucciK/package-manager/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type zipHandler struct {
	wr *zip.Writer
}

func NewZipHandler(zipFile *os.File) ZipHandler {
	return &zipHandler{
		wr: zip.NewWriter(zipFile),
	}
}

func (z *zipHandler) GetWriter() *zip.Writer {
	return z.wr
}

func (z *zipHandler) WriteFile(file *os.File, subDir string) error {
	wr, err := z.wr.Create(filepath.Join(subDir, filepath.Base(file.Name())))
	if err != nil {
		return err
	}

	if _, err := io.Copy(wr, file); err != nil {
		return err
	}

	return nil
}

func (z *zipHandler) CloseWriter() error {
	return z.wr.Close()
}

func (z *zipHandler) Unzip(dst, zipPath string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("can't open zip reader: %w", err)
	}

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return errors.ErrInvalidPath
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return fmt.Errorf("can't create sub dir: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("can't create sub dir: %w", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("can't open dst file: %w", err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf("can't open file in archive: %w", err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return fmt.Errorf("can't copy from archive file to dst: %w", err)
		}

		if err := dstFile.Close(); err != nil {
			return fmt.Errorf("can't close dst file: %w", err)
		}

		if err := fileInArchive.Close(); err != nil {
			return fmt.Errorf("can't close archive file: %w", err)
		}
	}

	return nil
}
