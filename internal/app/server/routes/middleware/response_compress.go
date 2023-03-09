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

// Допустимые статусы и Content-Type
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

// minBytesToCompress - минимальный размер тела для подключения сжатия.
const minBytesToCompress = 10

// gzipWriter - middleware для сжатия ответа
type gzipWriter struct {
	http.ResponseWriter // встраиваем ResponseWriter и забираем методы
	Writer              io.Writer
	statusCode          int
	buf                 *bytes.Buffer
}

// Write - переопределенный метод ResponseWriter. Накапливает
// сообщения от последующих обработчиков в буфере.
func (gzw *gzipWriter) Write(b []byte) (int, error) {
	return gzw.buf.Write(b)
}

// writeResponse выявляет необходимость в сжатии и подменяет
// gzw.Writer и необходимости.
func (gzw *gzipWriter) writeResponse() {
	// если кодировать нужно, то создаем gzip.Writer,
	// заменяем дефолтный Writer и
	// указываем заголовок Content-Encoding
	if gzw.IsCompressableStatus() && gzw.IsCompressableContent() &&
		gzw.IsCompressableSize(gzw.buf.Bytes()) {
		gz, _ := gzip.NewWriterLevel(gzw.Writer, gzip.BestSpeed)
		gzw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		gzw.Writer = gz
	}
	gzw.ResponseWriter.WriteHeader(gzw.statusCode)
	gzw.Writer.Write(gzw.buf.Bytes())
}

// WriteHeader = переопределенный метод ResponseWriter: проверяет, нужно ли
// кодировать данные на основе statusCode и contentType.
//
// я хочу добавить проверки на тип контента и размер
// я вижу так: мне нужно вкрячиться куда-то после вызовов
// (может быть несколько!) w.Write, которые будут в самих обработчиках
// и при этом нужно сделать так, чтобы код самих обработчиков не надо было переписывать
// лучший вариант, который я смог придумать - переопределить WriteHeader и Write...
func (gzw *gzipWriter) WriteHeader(statusCode int) {
	// запомним statusCode
	gzw.statusCode = statusCode
	// проблема: я хочу добавить проверку:
	// если размер данных мал, то сжимать не нужно.
	// если я здесь вызову gzw.ResponseWriter.WriteHeader
	// то заголовки зафиксируются и я уже ничего не смогу изменить
	// а размер данных я пойму только во время вызова Write
	// поэтому отложим вызов gzw.ResponseWriter.WriteHeader до тех пор
}

// IsCompressableContent проверяет, имеет ли смысл сжимать данные на основе content type.
func (gzw *gzipWriter) IsCompressableContent() bool {
	ct := gzw.ResponseWriter.Header().Get("Content-type")
	for _, cct := range compressableContentTypes {
		if strings.Contains(ct, cct) {
			return true
		}
	}
	return false
}

// IsCompressableSize проверяет, имеет ли смысл сжимать данные на основе их размера.
func (gzw *gzipWriter) IsCompressableSize(b []byte) bool {
	return binary.Size(b) >= minBytesToCompress
}

// IsCompressableStatus проверяет, имеет ли смысл сжимать данные на основе статуса
// ошибки, редиректы и т.д. нужно просто скипать.
func (gzw *gzipWriter) IsCompressableStatus() bool {
	return slices.Contains(statusesToCompress, gzw.statusCode)
}

// Close корректно закрывает gzip.Writer.
//
// Writer хочется не забыть закрыть, но эти вызовы нужно отложить
// до завершения ResponseGZipCompess, иначе не получится записать ответ
func (gzw *gzipWriter) Close() {
	w, ok := gzw.Writer.(*gzip.Writer)
	if !ok {
		return
	}
	w.Flush()
	w.Close()
}

// NewGzipWriter - конструктор gzipWriter.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	buf := new(bytes.Buffer)
	return &gzipWriter{ResponseWriter: w, Writer: w, buf: buf}
}

// ResponseGZipCompess возвращает обработчик middleware.
func ResponseGZipCompess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// если клиент не умеет декодировать gzip, не будем кодировать
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// создаем кастомный ResponseWriter и передаем его дальше для
		// выполнения запроса
		writer := NewGzipWriter(w)
		defer writer.Close()
		next.ServeHTTP(writer, r)
		writer.writeResponse()
	})
}
