package ccdb

import (
	"crypto/tls"
	"io"
)

func UploadFile(uploadSubdirUrl string, tlsConfig *tls.Config, sor, eor uint64, filename string, file io.Reader) error {
	return uploadFile(filename, uploadSubdirUrl, file, sor, eor, tlsConfig)
}
