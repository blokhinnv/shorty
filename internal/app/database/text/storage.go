// Пакет для создания текстового хранилища URL
package text

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// TextStorage реализует интерфейс Storage на основе текстового файла.
type TextStorage struct {
	filePath  string
	ttlOnDisk time.Duration
	ttlInMem  time.Duration
	db        []storage.Record
	toUpdate  map[string]time.Time
	buf       *bytes.Buffer
	encoder   *json.Encoder
	mu        sync.Mutex
}

// Настройки для выборки данных из текстового файла.
const (
	ByUserID = iota
	ByURLID
	ByURL
	ByUserIDAndURLID
)

// TextStorageRequest - структура для оформления запроса к хранилищу.
type TextStorageRequest struct {
	URL    string
	URLID  string
	UserID uint32
	Size   int
	How    int
	URLIDs []string
}

// NewTextStorage - конструктор нового хранилища URL.
func NewTextStorage(conf *TextStorageConfig) (*TextStorage, error) {
	if conf.ClearOnStart {
		os.Remove(conf.FileStoragePath)
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	s := &TextStorage{
		filePath:  conf.FileStoragePath,
		ttlOnDisk: conf.TTLOnDisk,
		ttlInMem:  conf.TTLInMemory,
		toUpdate:  make(map[string]time.Time),
		buf:       buf,
		encoder:   json.NewEncoder(buf),
	}
	file, err := os.OpenFile(s.filePath, os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	s.registerUpdateStorage()
	defer file.Close()
	return s, nil
}

// -------- Логика для обновления хранилища ----------

// UpdateStorage обновляет файл хранилища: удаляет старые URL
// и обновляет информацию о последнем запросе
func (s *TextStorage) UpdateStorage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Infof("Updating disk storage...\n")

	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}

	// читаем с диска, выбрасываем старье
	newDB := make([]storage.Record, 0)
	var r storage.Record
	decoder := json.NewDecoder(file)
	for decoder.More() {
		err = decoder.Decode(&r)
		if err != nil {
			log.Fatal(err)
		}
		if time.Since(r.Added) < s.ttlOnDisk {
			if reqTime, ok := s.toUpdate[r.URL]; ok {
				r.RequestedAt = reqTime
				log.Infof("Updated last request time of %+v \n", r)
				delete(s.toUpdate, r.URL)
			}
			newDB = append(newDB, r)
		} else {
			log.Infof("Removing %+v from disk \n", r)
		}
	}
	file.Close()
	// записываем свежую версию хранилища на диск
	os.Remove(s.filePath)
	file, err = os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, r := range newDB {
		err = encoder.Encode(r)
		if err != nil {
			panic(err)
		}
	}
	// обновим хранилище в памяти
	s.db = newDB
}

// registerUpdateStorage запускает обновление файла по таймеру.
func (s *TextStorage) registerUpdateStorage() {
	ticker := time.NewTicker(s.ttlOnDisk)
	go func() {
		for range ticker.C {
			s.UpdateStorage()
		}
	}()
}

// DeleteNotRequested удаляет из памяти те URL, которые запрашивались давно.
func (s *TextStorage) DeleteNotRequested() {
	filtered := make([]storage.Record, 0)
	for _, rec := range s.db {
		if time.Since(rec.RequestedAt) < s.ttlInMem {
			filtered = append(filtered, rec)
		} else {
			log.Infof("Remove %+v from memory\n", rec)
		}
	}
	s.db = filtered
}

// AppendFromBuffer обновляет файл информацией из буфера.
func (s *TextStorage) AppendFromBuffer() error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	// запишем в файл
	_, err = file.Write(s.buf.Bytes())
	if err != nil {
		return err
	}
	s.buf.Reset()
	return nil
}

// UpdateFile обновляет файл на основе мапы записей.
func (s *TextStorage) UpdateFile(newRecords map[string]storage.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// пометить в файле как удаленные
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}

	var rec storage.Record
	decoder := json.NewDecoder(file)
	for decoder.More() {
		err = decoder.Decode(&rec)
		if err != nil {
			log.Fatal(err)
		}
		if newRec, ok := newRecords[rec.URL]; ok {
			err = s.encoder.Encode(newRec)
		} else {
			err = s.encoder.Encode(rec)
		}
		if err != nil {
			return err
		}
	}
	file.Close()
	// удаляем файл и создаем новый
	err = os.Remove(s.filePath)
	if err != nil {
		return err
	}
	file, err = os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(s.buf.Bytes())
	if err != nil {
		return err
	}
	s.buf.Reset()
	return nil
}

// FindInMem ищет в памяти URL по urlID.
func (s *TextStorage) FindInMem(request TextStorageRequest) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	for _, rec := range s.db {
		matchURLID := request.How == ByURLID && rec.URLID == request.URLID
		matchUserID := request.How == ByUserID && rec.UserID == request.UserID
		matchURL := request.How == ByURL && rec.URL == request.URL
		// в памяти никогда не будет isDeleted записей
		// (они удаляются в DeleteMany)
		// но на всякий случай проверим
		if (matchURLID || matchUserID || matchURL) && !rec.IsDeleted {
			rec.RequestedAt = time.Now()
			s.toUpdate[rec.URL] = time.Now()
			results = append(results, rec)
		}
		if request.Size > 0 && len(results) == request.Size {
			return results, nil
		}
	}
	if len(results) == 0 {
		return results, storage.ErrURLWasNotFound
	}
	return results, nil
}

// FindInFile ищет в файле URL по urlID.
func (s *TextStorage) FindInFile(request TextStorageRequest) ([]storage.Record, error) {
	results := make([]storage.Record, 0)
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		return results, err
	}
	defer file.Close()

	var rec storage.Record
	decoder := json.NewDecoder(file)
	for decoder.More() {
		err = decoder.Decode(&rec)
		if err != nil {
			log.Fatal(err)
		}
		matchURLID := request.How == ByURLID && rec.URLID == request.URLID
		matchUserID := request.How == ByUserID && rec.UserID == request.UserID
		matchURL := request.How == ByURL && rec.URL == request.URL
		matchAnyURLIDAndUserID := request.How == ByUserIDAndURLID && request.URLIDs != nil &&
			rec.UserID == request.UserID &&
			slices.Contains(request.URLIDs, rec.URLID)
		if matchURLID || matchUserID || matchURL || matchAnyURLIDAndUserID {
			s.toUpdate[rec.URL] = time.Now()
			results = append(results, rec)
		}
		if request.Size > 0 && len(results) == request.Size {
			return results, nil
		}
	}
	if len(results) == 0 {
		return results, storage.ErrURLWasNotFound
	}
	return results, nil
}

// ------ Реализация интерфейса Storage ---------

// AddURL - метод для добавления нового URL в файл.
func (s *TextStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	// Попробуем найти запись в хранилище - если есть, то добавлять не надо
	req := TextStorageRequest{URL: url, Size: 1, How: ByURL}
	result, err := s.FindInFile(req)
	if err != nil && !errors.Is(err, storage.ErrURLWasNotFound) {
		return err
	}
	foundDeleted := make(map[string]storage.Record, 0)
	for _, rec := range result {
		// если нашли подходящий Url, у которого isDeleted=True, надо сбросить флаг
		if rec.IsDeleted {
			rec.IsDeleted = false
			foundDeleted[rec.URL] = rec
		}
	}
	// обновим файл
	s.UpdateFile(foundDeleted)
	// если нашли такую запись, вернем ошибку
	if len(result)-len(foundDeleted) > 0 {
		return fmt.Errorf(
			"%w: url=%v, urlID=%v, userID=%v",
			storage.ErrUniqueViolation,
			url,
			urlID,
			userID,
		)
	}

	r := storage.Record{
		URL:         url,
		URLID:       urlID,
		UserID:      userID,
		Added:       time.Now(),
		RequestedAt: time.Now(),
	}
	err = s.encoder.Encode(r)
	if err != nil {
		return err
	}
	log.Infof("Added %v=>%v to buffer\n", url, urlID)
	// добавим в память
	s.db = append(s.db, r)
	// добавим в файл
	err = s.AppendFromBuffer()
	if err != nil {
		return err
	}
	// почистим память от старых запросов
	s.DeleteNotRequested()
	return nil
}

// GetURLByID возвращает URL по его ID (сначала смотрит в памяти, потом в файле).
func (s *TextStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	req := TextStorageRequest{URLID: urlID, Size: 1, How: ByURLID}
	r, err := s.FindInMem(req)
	if errors.Is(err, storage.ErrURLWasNotFound) {
		r, err = s.FindInFile(req)
	}
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	rec := r[0]
	if rec.IsDeleted {
		return storage.Record{}, storage.ErrURLWasDeleted
	}
	return rec, nil
}

// GetURLsByUser получает URLs по ID пользователя.
func (s *TextStorage) GetURLsByUser(ctx context.Context, userID uint32) ([]storage.Record, error) {
	// Смотрим только в файле, т.к. его все равно придется
	// смотреть, чтобы быть уверенным, что нашли все.
	req := TextStorageRequest{UserID: userID, Size: 0, How: ByUserID}
	rFile, err := s.FindInFile(req)
	if err != nil {
		return nil, err
	}
	result := make([]storage.Record, 0)
	for _, rec := range rFile {
		if !rec.IsDeleted {
			result = append(result, rec)
		}
	}
	return result, nil
}

// AddURLBatch добавляет пакет URLов в хранилище.
func (s *TextStorage) AddURLBatch(
	ctx context.Context,
	urlIDs map[string]string,
	userID uint32,
) error {
	// Попробуем найти запись в хранилище - если есть, то добавлять не надо
	var violationErr error
	foundDeleted := make(map[string]storage.Record, 0)
	for url, urlID := range urlIDs {
		req := TextStorageRequest{URL: url, Size: 1, How: ByURL}
		result, err := s.FindInFile(req)
		if err != nil && !errors.Is(err, storage.ErrURLWasNotFound) {
			return err
		}
		// не нашли такой url: надо добавить
		if len(result) == 0 {
			r := storage.Record{
				URL:         url,
				URLID:       urlID,
				UserID:      userID,
				Added:       time.Now(),
				RequestedAt: time.Now(),
			}
			// добавим в память
			s.db = append(s.db, r)
			err = s.encoder.Encode(r)
			if err != nil {
				return err
			}
			log.Infof("Added %v=>%v to buffer\n", url, urlID)
			continue
		}
		// нашли такой урл
		rec := result[0]
		if rec.IsDeleted {
			// он удален => надо отметить как не удаленный
			rec.IsDeleted = false
			foundDeleted[rec.URL] = rec
		} else {
			// он не удален => надо сообщить о дубле
			violationErr = fmt.Errorf(
				"%w: url=%v, urlID=%v, userID=%v",
				storage.ErrUniqueViolation,
				url,
				urlID,
				userID,
			)
		}
	}
	s.UpdateFile(foundDeleted)
	// добавим в файл
	s.AppendFromBuffer()
	// почистим память от старых запросов
	s.DeleteNotRequested()
	if violationErr != nil {
		return violationErr
	}
	return nil
}

// DeleteMany устанавливает отметку об удалении URL.
func (s *TextStorage) DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error {
	// удалить из памяти
	newMem := make([]storage.Record, 0, len(s.db))
	for _, rec := range s.db {
		if !(rec.UserID == userID && slices.Contains(urlIDs, rec.URLID)) {
			newMem = append(newMem, rec)
		}
	}
	s.db = newMem
	req := TextStorageRequest{UserID: userID, URLIDs: urlIDs, How: ByUserIDAndURLID}
	result, err := s.FindInFile(req)
	if err != nil {
		if errors.Is(err, storage.ErrURLWasNotFound) {
			return nil
		}
		return err
	}

	toDelete := make(map[string]storage.Record, 0)
	for _, rec := range result {
		rec.IsDeleted = true
		toDelete[rec.URL] = rec
	}
	err = s.UpdateFile(toDelete)
	if err != nil {
		return err
	}
	return nil
}

// Close закрывает соединение с хранилищем
func (s *TextStorage) Close(ctx context.Context) {
}

// Ping проверяет соединение с хранилищем.
func (s *TextStorage) Ping(ctx context.Context) bool {
	_, err := os.Stat(s.filePath)
	return err == nil
}

// Clear очищает хранилище.
func (s *TextStorage) Clear(ctx context.Context) error {
	s.db = s.db[:0]
	err := os.Remove(s.filePath)
	if err != nil {
		return err
	}
	_, err = os.Create(s.filePath)
	if err != nil {
		return err
	}
	return nil
}
