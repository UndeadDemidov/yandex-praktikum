package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

// Storage реализует хранение ссылок в файле.
// Выполнена простейшая реализация для сдачи работы.
type Storage struct {
	mx sync.Mutex
	// Ридер один, но в теории правильней было бы сделать пул ридеров,
	// так как в таком сервисе кол-во чтений в разы (десятки/сотни раз) больше,
	// чем записей
	storageReader *Reader
	storageWriter *Writer
}

var _ handlers.Repository = (*Storage)(nil)

// NewStorage cоздаёт и возвращает экземпляр Storage
func NewStorage(filename string) (fs *Storage, err error) {
	if err = utils.CheckFilename(filename); err != nil {
		return nil, err
	}
	fs = &Storage{}
	fs.storageReader, err = NewReader(filename)
	if err != nil {
		return nil, err
	}
	fs.storageWriter, err = NewWriter(filename)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// isExist проверяет наличие в файле указанного ID
// Если такой ID входит как подстрока в ссылку, то результат будет такой же, как если бы был найден ID
func (s *Storage) isExist(_ context.Context, id string) bool {
	_, err := s.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(s.storageReader.file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), id) {
			// Не обрабатывается ситуация, когда в одной из ссылок может быть подстрока равная ID
			// Для этого можно сделать decoding JSON или захардкодить `"Key:"id"`
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}
	return false
}

// Store - сохраняет ID и ссылку в формате JSON во внешнем файле
func (s *Storage) Store(ctx context.Context, user string, link string) (id string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	id, err = utils.CreateShortID(ctx, s.isExist)
	if err != nil {
		return "", err
	}

	err = s.store(user, id, link)
	if err != nil {
		return "", err
	}

	return id, err
}

func (s *Storage) store(user string, id string, link string) error {
	a := Alias{User: user, Key: id, URL: link}
	err := s.storageWriter.Write(&a)
	if err != nil {
		return err
	}
	return nil
}

// Restore - находит по ID ссылку во внешнем файле, где данные хранятся в формате JSON
func (s *Storage) Restore(_ context.Context, id string) (link string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err = s.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	for {
		alias, err := s.storageReader.Read()
		if err != nil {
			return "", fmt.Errorf(storages.ErrLinkNotFound, id)
		}

		if alias.Key == id {
			return alias.URL, nil
		}
	}
}

// Unstore - помечает список ранее сохраненных ссылок удаленными
// только тех ссылок, которые принадлежат пользователю
// Только для совместимости контракта
func (s *Storage) Unstore(_ context.Context, _ string, _ []string) {
	panic("not implemented for file storage")
}

// GetUserStorage возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (s *Storage) GetUserStorage(_ context.Context, user string) map[string]string {
	s.mx.Lock()
	defer s.mx.Unlock()

	m := map[string]string{}
	_, err := s.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return map[string]string{}
	}

	scanner := bufio.NewScanner(s.storageReader.file)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, user) {
			alias := &Alias{}
			dec := json.NewDecoder(bytes.NewBufferString(txt))
			if err := dec.Decode(&alias); err != nil {
				return map[string]string{}
			}
			m[alias.Key] = alias.URL
		}
	}

	if err := scanner.Err(); err != nil {
		return map[string]string{}
	}

	return m
}

// StoreBatch сохраняет пакет ссылок из map[correlation_id]original_link и возвращает map[correlation_id]short_link
func (s *Storage) StoreBatch(ctx context.Context, user string, batchIn map[string]string) (batchOut map[string]string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	batchOut = make(map[string]string)
	var id string
	for corrID, link := range batchIn {
		id, err = utils.CreateShortID(ctx, s.isExist)
		if err != nil {
			return nil, err
		}
		err = s.store(user, id, link)
		if err != nil {
			return nil, err
		}
		batchOut[corrID] = id
	}
	return batchOut, nil
}

// Ping проверяет, что файл хранения доступен и экземпляры инициализированы
func (s *Storage) Ping(_ context.Context) error {
	_, err := s.storageWriter.file.Stat()
	if err != nil {
		return storages.ErrStorageIsUnavailable
	}
	_, err = s.storageReader.file.Stat()
	if err != nil {
		return storages.ErrStorageIsUnavailable
	}
	return nil
}

// Close закрывает все файлы, открытых для записи и чтения
func (s *Storage) Close() error {
	err1 := s.storageReader.Close()
	err2 := s.storageWriter.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

type Writer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewWriter(fileName string) (*Writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &Writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Writer) Write(event *Alias) error {
	return p.encoder.Encode(&event)
}

func (p *Writer) Close() error {
	return p.file.Close()
}

type Reader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewReader(fileName string) (*Reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &Reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Reader) Read() (*Alias, error) {
	alias := &Alias{}
	if err := c.decoder.Decode(&alias); err != nil {
		return nil, err
	}
	return alias, nil
}

func (c *Reader) Close() error {
	return c.file.Close()
}

// Alias - структура хранения ID и URL во внешнем файле
type Alias struct {
	User string
	Key  string
	URL  string
}
