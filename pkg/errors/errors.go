package errors

import "errors"

var (
	ErrNoArgs            = errors.New("no args passed")
	ErrInvalidVerPrefix  = errors.New("invalid version prefix")
	ErrZeroBytesWriten   = errors.New("zero bytes written to remote file")
	ErrNilDownloadedFile = errors.New("downloaded file is nil")
	ErrInvalidPath       = errors.New("invalid file path")
)
