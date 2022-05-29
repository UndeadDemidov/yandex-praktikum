package storages

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
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

// FileStorage реализует хранение ссылок в файле.
// Выполнена простейшая реализация для сдачи работы.
type FileStorage struct {
	mx sync.Mutex
	// Ридер один, но в теории правильней было бы сделать пул ридеров,
	// так как в таком сервисе кол-во чтений в разы (десятки/сотни раз) больше,
	// чем записей
	storageReader *reader
	storageWriter *writer
}

var _ handlers.Repository = (*FileStorage)(nil)

// NewFileStorage cоздаёт и возвращает экземпляр FileStorage
func NewFileStorage(filename string) (fs *FileStorage, err error) {
	if err = utils.CheckFilename(filename); err != nil {
		return nil, err
	}
	fs = &FileStorage{}
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

// IsExist проверяет наличие в файле указанного ID
// Если такой ID входит как подстрока в ссылку, то результат будет такой же, как если бы был найден ID
func (f *FileStorage) IsExist(_ context.Context, id string) bool {
	f.mx.Lock()
	defer f.mx.Unlock()

	_, err := f.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(f.storageReader.file)
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
func (f *FileStorage) Store(_ context.Context, user string, id string, link string) error {
	f.mx.Lock()
	defer f.mx.Unlock()

	return f.store(user, id, link)
}

func (f *FileStorage) store(user string, id string, link string) error {
	a := Alias{User: user, Key: id, URL: link}
	err := f.storageWriter.Write(&a)
	if err != nil {
		return err
	}
	return nil
}

// Restore - находит по ID ссылку во внешнем файле, где данные хранятся в формате JSON
func (f *FileStorage) Restore(_ context.Context, id string) (link string, err error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	_, err = f.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	for {
		alias, err := f.storageReader.Read()
		if err != nil {
			return "", fmt.Errorf(ErrLinkNotFound, id)
		}

		if alias.Key == id {
			return alias.URL, nil
		}
	}
}

// GetAllUserLinks возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (f *FileStorage) GetAllUserLinks(_ context.Context, user string) map[string]string {
	f.mx.Lock()
	defer f.mx.Unlock()

	m := map[string]string{}
	_, err := f.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return map[string]string{}
	}

	scanner := bufio.NewScanner(f.storageReader.file)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, user) {
			alias := &Alias{}
			// ToDo вынести декодеры из reader/writer? Что-то не режется красиво на слои
			// Либо опять по всему файлу бежать с unmarshal
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

// StoreBatch сохраняет пакет ссылок из map[id]link
func (f *FileStorage) StoreBatch(_ context.Context, user string, batch map[string]string) error {
	f.mx.Lock()
	defer f.mx.Unlock()

	for k, v := range batch {
		err := f.store(user, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close закрывает все файлы, открытых для записи и чтения
func (f *FileStorage) Close() error {
	err1 := f.storageReader.Close()
	err2 := f.storageWriter.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

type writer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewWriter(fileName string) (*writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *writer) Write(event *Alias) error {
	return p.encoder.Encode(&event)
}

func (p *writer) Close() error {
	return p.file.Close()
}

type reader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewReader(fileName string) (*reader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *reader) Read() (*Alias, error) {
	alias := &Alias{}
	if err := c.decoder.Decode(&alias); err != nil {
		return nil, err
	}
	return alias, nil
}

func (c *reader) Close() error {
	return c.file.Close()
}

// Alias - структура хранения ID и URL во внешнем файле
type Alias struct {
	User string
	Key  string
	URL  string
}
