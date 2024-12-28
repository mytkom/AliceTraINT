package ccdb

import (
	"crypto/tls"
	"io"
)

func UploadFile(uploadSubdirUrl string, cert *tls.Certificate, sor, eor uint64, filename string, file io.Reader) error {
	ssl := &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
	}

	err := uploadFile(filename, uploadSubdirUrl, file, sor, eor, ssl)
	if err != nil {
		return err
	}

	return nil
}
