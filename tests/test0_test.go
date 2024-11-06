package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/coldstar-507/media-server/internal/handlers"
	"github.com/coldstar-507/media-server/internal/paths"
)

var (
	id        = "abc"
	permanent = false
	// file format -> [2 bytes for metadata len, metadata, media]
	content  = []byte{0, 6, 0, 1, 2, 3, 4, 5, 10, 11, 12, 13}
	metadata = content[2 : 2+6]
	data     = content[2+6:]
	realHex  = "000001914cd0eca80100000001914cd0bf7e9e62823103e20400000007010d"
)

func TestMain(m *testing.M) {
	paths.InitWD("/home/scott/dev/down4/backend/media-server")
	code := m.Run()
	os.Exit(code)
}

func TestCheckIfPermanent(t *testing.T) {
	if !paths.IsPermanent(realHex) {
		t.Error("should be permanent")
	}
}

func TestHandleWriteMedia(t *testing.T) {
	val := bytes.NewReader(content)
	if err := handlers.WriteMedia(id, permanent, val); err != nil {
		t.Error("TestHandleWriteMedia error writing media :", err)
	}
}

func TestHandleStreamMedia(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.StreamMedia(id, permanent, buf); err != nil {
		t.Error("TestHandleStreamMedia error streaming media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("TestHandleReadMedia error: expected=%v, got=%v\n", data, buf.Bytes())
	}
}

func TestHandleReadMetadata(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.ReadMetadata(id, permanent, buf); err != nil {
		t.Error("TestHandleReadMetadata error streaming media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), metadata) {
		t.Errorf("TestHandleReadMetadata error: expected=%v, got=%v\n", metadata, buf.Bytes())
	}
}

func TestHandleReadMedia(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 20))
	if err := handlers.ReadMedia(id, permanent, buf); err != nil {
		t.Error("TestHandleReadMedia error reading media: ", err)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("TestHandleReadMedia error: expected=%v, got=%v\n", content, buf.Bytes())
	}
}

func TestRemoveMedia(t *testing.T) {
	if err := handlers.RemoveMedia(id, permanent); err != nil {
		t.Error("TestRemoveMedia error:", err)
	}
}
