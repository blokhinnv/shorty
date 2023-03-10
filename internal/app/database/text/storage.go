// Package text implements Text-based storage.
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

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"golang.org/x/exp/slices"
)

// TextStorage implements the Storage interface based on a text file.
type TextStorage struct {
	filePath  string
	ttlOnDisk time.Duration
	ttlInMem  time.Duration
	db        []storage.Record
	toUpdate  map[string]time.Time
	buf       *bytes.Buffer
	encoder   *json.Encoder
	mu        sync.Mutex
	quit      chan struct{}
}

// Settings for fetching data from a text file.
const (
	ByUserID = iota
	ByURLID
	ByURL
	ByUserIDAndURLID
)

// TextStorageRequest - a structure for making a request to the storage.
type TextStorageRequest struct {
	URL    string
	URLID  string
	UserID uint32
	Size   int
	How    int
	URLIDs []string
}

// NewTextStorage - constructor for a new URL storage.
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
		quit:      make(chan struct{}),
	}
	file, err := os.OpenFile(s.filePath, os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	s.registerUpdateStorage()
	defer file.Close()
	return s, nil
}

// -------- Logic for updating storage ----------

// UpdateStorage updates the storage file: removes old URLs
// and update information about the last request
func (s *TextStorage) UpdateStorage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Infof("Updating disk storage...\n")

	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatalf("unable to open file: %v", err)
	}

	// read from disk, discard junk
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
	// write the latest version of the repository to disk
	os.Remove(s.filePath)
	file, err = os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatalf("unable to reopen file: %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, r := range newDB {
		err = encoder.Encode(r)
		if err != nil {
			panic(err)
		}
	}
	// update storage in memory
	s.db = newDB
}

// registerUpdateStorage starts a file update on a timer.
func (s *TextStorage) registerUpdateStorage() {
	ticker := time.NewTicker(s.ttlOnDisk)
	go func() {
		// for range ticker.C {
		// s.UpdateStorage()
		// }
		for {
			select {
			case <-ticker.C:
				s.UpdateStorage()
			case <-s.quit:
				return
			}
		}
	}()
}

// DeleteNotRequested removes URLs from memory that were requested a long time ago.
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

// AppendFromBuffer updates the file with information from the buffer.
func (s *TextStorage) AppendFromBuffer() error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	// write to file
	_, err = file.Write(s.buf.Bytes())
	if err != nil {
		return err
	}
	s.buf.Reset()
	return nil
}

// UpdateFile updates the file based on the entry map.
func (s *TextStorage) UpdateFile(newRecords map[string]storage.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// mark file as deleted
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
	// delete the file and create a new one
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

// FindInMem searches memory for a URL by urlID.
func (s *TextStorage) FindInMem(request TextStorageRequest) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	for _, rec := range s.db {
		matchURLID := request.How == ByURLID && rec.URLID == request.URLID
		matchUserID := request.How == ByUserID && rec.UserID == request.UserID
		matchURL := request.How == ByURL && rec.URL == request.URL
		// there will never be isDeleted records in memory
		// (they are deleted in DeleteMany)
		// but just in case, check
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

// FindInFile searches a file for a URL by urlID.
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

// ------ Implementation of the Storage interface ---------

// AddURL - method for adding a new URL to the file.
func (s *TextStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	// Let's try to find a record in the storage - if there is, then do not add
	req := TextStorageRequest{URL: url, Size: 1, How: ByURL}
	result, err := s.FindInFile(req)
	if err != nil && !errors.Is(err, storage.ErrURLWasNotFound) {
		return err
	}
	foundDeleted := make(map[string]storage.Record, 0)
	for _, rec := range result {
		// if you find a suitable Url that has isDeleted=True, you need to reset the flag
		if rec.IsDeleted {
			rec.IsDeleted = false
			foundDeleted[rec.URL] = rec
		}
	}
	// update the file
	s.UpdateFile(foundDeleted)
	// if such an entry is found, return an error
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
	// add to memory
	s.db = append(s.db, r)
	// add to file
	err = s.AppendFromBuffer()
	if err != nil {
		return err
	}
	// clear memory from old requests
	s.DeleteNotRequested()
	return nil
}

// GetURLByID returns a URL by its ID (first looks in memory, then in a file).
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

// GetURLsByUser gets URLs by user ID.
func (s *TextStorage) GetURLsByUser(ctx context.Context, userID uint32) ([]storage.Record, error) {
	// We look only in the file, because it will still have to
	// watch to make sure you've found everything.
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

// AddURLBatch adds a batch of URLs to the store.
func (s *TextStorage) AddURLBatch(
	ctx context.Context,
	urlIDs map[string]string,
	userID uint32,
) error {
	// Let's try to find a record in the storage - if there is, then do not add
	var violationErr error
	foundDeleted := make(map[string]storage.Record, 0)
	for url, urlID := range urlIDs {
		req := TextStorageRequest{URL: url, Size: 1, How: ByURL}
		result, err := s.FindInFile(req)
		if err != nil && !errors.Is(err, storage.ErrURLWasNotFound) {
			return err
		}
		// no such url found: add
		if len(result) == 0 {
			r := storage.Record{
				URL:         url,
				URLID:       urlID,
				UserID:      userID,
				Added:       time.Now(),
				RequestedAt: time.Now(),
			}
			// add to memory
			s.db = append(s.db, r)
			err = s.encoder.Encode(r)
			if err != nil {
				return err
			}
			log.Infof("Added %v=>%v to buffer\n", url, urlID)
			continue
		}
		// found this url
		rec := result[0]
		if rec.IsDeleted {
			// it's deleted => should be marked as not deleted
			rec.IsDeleted = false
			foundDeleted[rec.URL] = rec
		} else {
			// it's not deleted => need to report a duplicate
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
	// add to file
	s.AppendFromBuffer()
	// clear memory from old requests
	s.DeleteNotRequested()
	if violationErr != nil {
		return violationErr
	}
	return nil
}

// DeleteMany flags the URL to be deleted.
func (s *TextStorage) DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error {
	// remove from memory
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

// Close closes the connection to the store
func (s *TextStorage) Close(ctx context.Context) {
	s.quit <- struct{}{}
}

// Ping checks the connection to the repository.
func (s *TextStorage) Ping(ctx context.Context) bool {
	_, err := os.Stat(s.filePath)
	return err == nil
}

// Clear clears the storage.
func (s *TextStorage) Clear(ctx context.Context) error {
	s.db = s.db[:0]
	err := os.Remove(s.filePath)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.filePath, os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
