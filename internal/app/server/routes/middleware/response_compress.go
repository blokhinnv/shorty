package middleware

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"
)

var statusesToCompress = []int{
	http.StatusCreated,
	http.StatusOK,
}
var compressableContentTypes = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

const minBytesToCompress = 1024

type gzipWriter struct {
	http.ResponseWriter // встраиваем ResponseWriter и забираем методы
	Writer              io.Writer
	statusCode          int
}

// Переопределенный метод ResponseWriter: проверяет, нужно ли
// кодировать данные на основе их размера, отправляет заголовки и пишет ответ
func (gzw *gzipWriter) Write(b []byte) (int, error) {
	if gzw.IsCompressableStatus() && gzw.IsCompressableContent() && gzw.IsCompressableSize(b) {
		gzw.SwitchToCompressMode()
	} else {
		gzw.ResponseWriter.WriteHeader(gzw.statusCode)
	}
	return gzw.Writer.Write(b)
}

// Создает gzip.Writer, заменяет дефолтный Writer и
// указывает заголовок Content-Encoding
func (gzw *gzipWriter) SwitchToCompressMode() {
	gz, _ := gzip.NewWriterLevel(gzw.Writer, gzip.BestSpeed)
	gzw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	gzw.Writer = gz
	gzw.ResponseWriter.WriteHeader(gzw.statusCode)
}

// Переопределенный метод ResponseWriter: проверяет, нужно ли
// кодировать данные на основе statusCode и contentType
// я хочу добавить проверки на тип контента и размер
// я вижу так: мне нужно вкрячиться куда-то между вызовами w.Header().Set
// и w.WriteHeader + w.Write, которые будут в самих обработчиках
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

// Проверяет, имеет ли смысл сжимать данные на основе content type
func (gzw *gzipWriter) IsCompressableContent() bool {
	ct := gzw.ResponseWriter.Header().Get("Content-type")
	for _, cct := range compressableContentTypes {
		if strings.Contains(ct, cct) {
			return true
		}
	}
	return false
}

// Проверяет, имеет ли смысл сжимать данные на основе их размера
func (gzw *gzipWriter) IsCompressableSize(b []byte) bool {
	return binary.Size(b) >= minBytesToCompress
}

// Проверяет, имеет ли смысл сжимать данные на основе статуса
// ошибки, редиректы и т.д. нужно просто скипать
func (gzw *gzipWriter) IsCompressableStatus() bool {
	return slices.Contains(statusesToCompress, gzw.statusCode)
}

// Корректно закрывает gzip.Writer
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

// Конструктор gzipWriter
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{ResponseWriter: w, Writer: w}
}

// Middleware для сжатия ответа
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
	})
}
