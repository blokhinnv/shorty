package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"
)

// Valid statuses and Content-Type
var (
	statusesToCompress = []int{
		http.StatusCreated,
		http.StatusOK,
	}
	compressableContentTypes = []string{
		"application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml",
	}
)

// minBytesToCompress - the minimum size of the body to enable compression.
const minBytesToCompress = 10

// gzipWriter - middleware for compressing the response
type gzipWriter struct {
	http.ResponseWriter // embed ResponseWriter and pick up methods
	Writer              io.Writer
	statusCode          int
	buf                 *bytes.Buffer
}

// Write is an overridden ResponseWriter method. accumulates
// messages from subsequent handlers in the buffer.
func (gzw *gzipWriter) Write(b []byte) (int, error) {
	return gzw.buf.Write(b)
}

// writeResponse detects the need for compression and replaces
// gzw.Writer and needed.
func (gzw *gzipWriter) writeResponse() {
	// if you need to encode, then create a gzip.Writer,
	// replace the default Writer and
	// specify the Content-Encoding header
	if gzw.isCompressableStatus() && gzw.isCompressableContent() &&
		gzw.isCompressableSize(gzw.buf.Bytes()) {
		gz, _ := gzip.NewWriterLevel(gzw.Writer, gzip.BestSpeed)
		gzw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		gzw.Writer = gz
	}
	gzw.ResponseWriter.WriteHeader(gzw.statusCode)
	gzw.Writer.Write(gzw.buf.Bytes())
}

// WriteHeader = overridden ResponseWriter method: checks if the
// encode data based on statusCode and contentType.
func (gzw *gzipWriter) WriteHeader(statusCode int) {
	// remember statusCode
	gzw.statusCode = statusCode
	// problem: I want to add a check:
	// if the data size is small, then no need to compress.
	// if I call gzw.ResponseWriter.WriteHeader here
	// then the headers will be fixed and I won't be able to change anything
	// and I will understand the size of the data only during the call to Write
	// so we'll defer calling gzw.ResponseWriter.WriteHeader until then
}

// isCompressableContent checks if it makes sense to compress data based on content type.
func (gzw *gzipWriter) isCompressableContent() bool {
	ct := gzw.ResponseWriter.Header().Get("Content-type")
	for _, cct := range compressableContentTypes {
		if strings.Contains(ct, cct) {
			return true
		}
	}
	return false
}

// isCompressableSize checks if it makes sense to compress the data based on its size.
func (gzw *gzipWriter) isCompressableSize(b []byte) bool {
	return binary.Size(b) >= minBytesToCompress
}

// isCompressableStatus checks if it makes sense to compress data based on status
// errors, redirects, etc. you just have to skim.
func (gzw *gzipWriter) isCompressableStatus() bool {
	return slices.Contains(statusesToCompress, gzw.statusCode)
}

// Close correctly closes gzip.Writer.
func (gzw *gzipWriter) Close() {
	// Writer wants to remember to close, but these calls should be deferred
	// before ResponseGZipCompess completes, otherwise it will not be possible to write the response
	w, ok := gzw.Writer.(*gzip.Writer)
	if !ok {
		return
	}
	w.Flush()
	w.Close()
}

// NewGzipWriter - gzipWriter constructor.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	buf := new(bytes.Buffer)
	return &gzipWriter{ResponseWriter: w, Writer: w, buf: buf}
}

// ResponseGZipCompess returns the middleware handler.
func ResponseGZipCompess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if client can't decode gzip, don't encode
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// create a custom ResponseWriter and pass it on to
		// request execution
		writer := NewGzipWriter(w)
		defer writer.Close()
		next.ServeHTTP(writer, r)
		writer.writeResponse()
	})
}
