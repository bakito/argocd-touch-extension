package server

import (
	"archive/tar"
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) handleExtensionTar(c *gin.Context) {
	archive, err := s.createTar([]byte("test"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "application/x-tar", archive)
}

func (s *server) createTar(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	hdr := &tar.Header{
		Name:    "resources/extension-touch.js",
		Mode:    0o644,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write(data); err != nil {
		return nil, err
	}

	_ = tw.Close()
	return buf.Bytes(), nil
}
