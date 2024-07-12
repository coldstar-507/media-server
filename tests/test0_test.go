package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/coldstar-507/media-server/internal/handlers"
	"github.com/coldstar-507/media-server/internal/paths"
)

var (
	id   = "abc"
	temp = true
	// file format -> [2 bytes for metadata len, metadata, media]
	content  = []byte{0, 6, 0, 1, 2, 3, 4, 5, 11, 12, 13, 14}
	metadata = content[2 : 2+6]
	data     = content[2+6:]
)

func TestMain(m *testing.M) {
	paths.InitWD(true)
	code := m.Run()
	os.Exit(code)
}

func TestHandleWriteMedia(t *testing.T) {
	val := bytes.NewReader(content)
	if err := handlers.WriteMedia(id, temp, val); err != nil {
		t.Error("TestHandleWriteMedia error writing media :", err)
	}
}

func TestHandleStreamMedia(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.StreamMedia(id, temp, buf); err != nil {
		t.Error("TestHandleStreamMedia error streaming media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("TestHandleReadMedia error: expected=%v, got=%v\n", data, buf.Bytes())
	}
}

func TestHandleReadMetadata(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.ReadMetadata(id, temp, buf); err != nil {
		t.Error("TestHandleReadMetadata error streaming media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), metadata) {
		t.Errorf("TestHandleReadMetadata error: expected=%v, got=%v\n", metadata, buf.Bytes())
	}
}

func TestHandleReadMedia(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.ReadMedia(id, temp, buf); err != nil {
		t.Error("TestHandleReadMedia error reading media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("TestHandleReadMedia error: expected=%v, got=%v\n", content, buf.Bytes())
	}
}

func TestRemoveMedia(t *testing.T) {
	if err := handlers.RemoveMedia(id, true); err != nil {
		t.Error("TestRemoveMedia error:", err)
	}
}
