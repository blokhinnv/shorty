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

type Record struct {
	URL         string    `json:"url"          valid:"url,required"`
	URLID       string    `json:"url_id"       valid:"url,required"`
	Added       time.Time `json:"added"`
	RequestedAt time.Time `json:"requested_at"`
}

type TextStorage struct {
	filePath  string
	ttlOnDisk time.Duration
	ttlInMem  time.Duration
	db        []Record
	toUpdate  map[string]time.Time
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

// Метод для добавления нового URL в файле
func (s *TextStorage) AddURL(url, urlID string) {
	// проблема: мне нужно открывать файл и на чтение, и на добавление
	// и при этом перемещать указатель то на начало, то на конец
	// но если открыть в режиме O_APPEND, то поведение Seek "is not specified" -- страшно
	// поэтому приходится открывать и закрывать файл
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// добавим на диск
	encoder := json.NewEncoder(file)
	r := Record{url, urlID, time.Now(), time.Now()}
	err = encoder.Encode(r)
	if err != nil {
		panic(err)
	}
	// добавим в память
	s.db = append(s.db, r)
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	// почистим память от старых запросов
	s.DeleteNotRequested()
}

// Удаляет из памяти те URL, которые запрашивались давно
func (s *TextStorage) DeleteNotRequested() {
	filtered := make([]Record, 0)
	for _, rec := range s.db {
		if time.Since(rec.RequestedAt) < s.ttlInMem {
			filtered = append(filtered, rec)
		} else {
			log.Printf("Remove %+v from memory\n", rec)
		}
	}
	s.db = filtered
}

// Ищет в памяти URL по urlID
func (s *TextStorage) FindInMem(urlID string) (Record, error) {
	for _, rec := range s.db {
		if rec.URLID == urlID {
			rec.RequestedAt = time.Now()
			s.toUpdate[rec.URL] = time.Now()
			return rec, nil
		}
	}
	return Record{}, storage.ErrURLWasNotFound
}

// Ищем в файле URL по urlID
func (s *TextStorage) FindInFile(urlID string) (Record, error) {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var r Record
	decoder := json.NewDecoder(file)
	for decoder.More() {
		err = decoder.Decode(&r)
		if err != nil {
			log.Fatal(err)
		}
		if r.URLID == urlID {
			s.toUpdate[r.URL] = time.Now()
			return r, nil
		}
	}
	return Record{}, storage.ErrURLWasNotFound
}

// Обновляет файл хранилища: удаляет старые URL и обновляет информацию
// о последнем запросе
func (s *TextStorage) UpdateStorage() {
	log.Printf("Updating disk storage...\n")

	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}

	// читаем с диска, выбрасываем старье
	newDB := make([]Record, 0)
	var r Record
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

func (s *TextStorage) registerUpdateStorage() {
	ticker := time.NewTicker(s.ttlOnDisk)
	go func() {
		for range ticker.C {
			s.UpdateStorage()
		}
	}()
}

// Возвращает URL по его ID (сначала смотрит в памяти, потом в файле)
func (s *TextStorage) GetURLByID(urlID string) (string, error) {
	r, err := s.FindInMem(urlID)
	if errors.Is(err, storage.ErrURLWasNotFound) {
		r, err = s.FindInFile(urlID)
	}
	if err != nil {
		return "", storage.ErrURLWasNotFound
	}

	return r.URL, nil
}

// Закрывает соединение с SQLite
func (s *TextStorage) Close() {
}
