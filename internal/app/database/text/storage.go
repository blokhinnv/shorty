// Пакет для создания текстового хранилища URL
package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

func init() {
	log.SetOutput(os.Stdout)
}

type TextStorage struct {
	filePath  string
	ttlOnDisk time.Duration
	ttlInMem  time.Duration
	db        []storage.Record
	toUpdate  map[string]time.Time
}

const (
	ByUserID = iota
	ByURLID
	ByBoth
)

type TextStorageRequest struct {
	URLID  string
	UserID uint32
	Size   int
	How    int
}

// Конструктор нового хранилища URL
func NewTextStorage(conf TextStorageConfig) (*TextStorage, error) {
	if conf.ClearOnStart {
		os.Remove(conf.FileStoragePath)
	}
	s := &TextStorage{
		filePath:  conf.FileStoragePath,
		ttlOnDisk: conf.TTLOnDisk,
		ttlInMem:  conf.TTLInMemory,
		toUpdate:  make(map[string]time.Time),
	}
	file, err := os.OpenFile(s.filePath, os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	s.registerUpdateStorage()
	defer file.Close()
	return s, nil
}

// -------- Логика для обновления хранилища ----------

// Обновляет файл хранилища: удаляет старые URL и обновляет информацию
// о последнем запросе
func (s *TextStorage) UpdateStorage() {
	log.Printf("Updating disk storage...\n")

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
				log.Printf("Updated last request time of %+v \n", r)
				delete(s.toUpdate, r.URL)
			}
			newDB = append(newDB, r)
		} else {
			log.Printf("Removing %+v from disk \n", r)
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

// Запускает обновление файла по таймеру
func (s *TextStorage) registerUpdateStorage() {
	ticker := time.NewTicker(s.ttlOnDisk)
	go func() {
		for range ticker.C {
			s.UpdateStorage()
		}
	}()
}

// Удаляет из памяти те URL, которые запрашивались давно
func (s *TextStorage) DeleteNotRequested() {
	filtered := make([]storage.Record, 0)
	for _, rec := range s.db {
		if time.Since(rec.RequestedAt) < s.ttlInMem {
			filtered = append(filtered, rec)
		} else {
			log.Printf("Remove %+v from memory\n", rec)
		}
	}
	s.db = filtered
}

// ------ Реализация интерфейса Storage ---------

// Метод для добавления нового URL в файле
func (s *TextStorage) AddURL(url, urlID string, userID uint32) error {
	// проблема: мне нужно открывать файл и на чтение, и на добавление
	// и при этом перемещать указатель то на начало, то на конец
	// но если открыть в режиме O_APPEND, то поведение Seek "is not specified" -- страшно
	// поэтому приходится открывать и закрывать файл
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	// Попробуем найти запись в хранилище - если есть, то добавлять не надо
	req := TextStorageRequest{URLID: urlID, UserID: userID, Size: 1, How: ByBoth}
	result, err := s.FetchFile(req)
	if err != nil && !errors.Is(err, storage.ErrURLWasNotFound) {
		return err
	}
	if len(result) > 0 {
		return nil
	}

	// добавим на диск
	encoder := json.NewEncoder(file)
	r := storage.Record{
		URL:         url,
		URLID:       urlID,
		UserID:      userID,
		Added:       time.Now(),
		RequestedAt: time.Now(),
	}
	err = encoder.Encode(r)
	if err != nil {
		return err
	}
	// добавим в память
	s.db = append(s.db, r)
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	// почистим память от старых запросов
	s.DeleteNotRequested()
	return nil
}

// Ищет в памяти URL по urlID
func (s *TextStorage) FetchMem(request TextStorageRequest) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	for _, rec := range s.db {
		matchURLID := request.How == ByURLID && rec.URLID == request.URLID
		matchUserID := request.How == ByUserID && rec.UserID == request.UserID
		matchBoth := request.How == ByBoth && rec.URLID == request.URLID &&
			rec.UserID == request.UserID
		if matchURLID || matchUserID || matchBoth {
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

// Ищем в файле URL по urlID
func (s *TextStorage) FetchFile(request TextStorageRequest) ([]storage.Record, error) {
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
		matchBoth := request.How == ByBoth && rec.URLID == request.URLID &&
			rec.UserID == request.UserID
		if matchURLID || matchUserID || matchBoth {
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

// Возвращает URL по его ID (сначала смотрит в памяти, потом в файле)
func (s *TextStorage) GetURLByID(urlID string) (storage.Record, error) {
	req := TextStorageRequest{URLID: urlID, Size: 1, How: ByURLID}
	r, err := s.FetchMem(req)
	if errors.Is(err, storage.ErrURLWasNotFound) {
		r, err = s.FetchFile(req)
	}
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}

	return r[0], nil
}

// Получает URLs по ID пользователя (смотрим только в файле, т.к. его все равно придется
// смотреть, чтобы быть уверенным, что нашли все)
func (s *TextStorage) GetURLsByUser(userID uint32) ([]storage.Record, error) {
	req := TextStorageRequest{UserID: userID, Size: 0, How: ByUserID}
	rFile, err := s.FetchFile(req)
	if err != nil {
		return nil, err
	}
	return rFile, nil
}

// Закрывает соединение с хранилищем
func (s *TextStorage) Close() {
}
