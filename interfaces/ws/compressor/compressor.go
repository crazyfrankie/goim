package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"

	"github.com/crazyfrankie/goim/pkg/errorx"
)

var (
	spWriter sync.Pool
	spReader sync.Pool
)

func init() {
	spWriter = sync.Pool{New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	}}
	spReader = sync.Pool{New: func() interface{} {
		return new(gzip.Reader)
	}}
}

type Compressor interface {
	Compress([]byte) ([]byte, error)
	Decompress([]byte) ([]byte, error)
}

func NewCompressor() Compressor {
	return &gzipCompressor{}
}

// gzipCompressor implements gzip compressor.
type gzipCompressor struct{}

func (g *gzipCompressor) Compress(data []byte) ([]byte, error) {
	writer := spWriter.Get().(*gzip.Writer)
	defer spWriter.Put(writer)

	var buf bytes.Buffer
	writer.Reset(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, errorx.Wrapf(err, "GzipCompressor.Compress: writing to gzip writer failed")
	}

	if err := writer.Close(); err != nil {
		return nil, errorx.Wrapf(err, "GzipCompressor.Compress: closing gzip writer failed")
	}

	return buf.Bytes(), nil
}

func (g *gzipCompressor) Decompress(data []byte) ([]byte, error) {
	reader := spReader.Get().(*gzip.Reader)
	defer spReader.Put(reader)

	if err := reader.Reset(bytes.NewReader(data)); err != nil {
		return nil, err
	}

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, errorx.Wrapf(err, "GzipCompressor.Decompress: reading from pooled gzip reader failed")
	}
	if err = reader.Close(); err != nil {
		return decompressedData, errorx.Wrapf(err, "GzipCompressor.Decompress: closing gzip reader failed")
	}
	return decompressedData, nil
}
