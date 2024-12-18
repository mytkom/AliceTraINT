package ccdb

import (
	"crypto/tls"
	"fmt"
	"io"

	"github.com/mytkom/AliceTraINT/internal/config"
)

func UploadFile(cfg *config.Config, sor, eor uint64, filename string, file io.Reader) error {
	cert, err := tls.LoadX509KeyPair(cfg.CCDBCertPath, cfg.CCDBKeyPath)
	if err != nil {
		return err
	}

	ssl := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	url := fmt.Sprintf("%s/%s", cfg.CCDBBaseURL, cfg.CCDBUploadSubdir)
	err = uploadFile(filename, url, file, sor, eor, ssl)
	if err != nil {
		return err
	}

	return nil
}
