package storages

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

// FileStorage реализует хранение ссылок в файле.
// Выполнена простейшая реализация для сдачи работы.
// Мой старик прежде чем покинуть этот говенный мир говорил:
// - Никогда не изобретай велосипед, все равно колеса круглее не будут.
// https://github.com/akrylysov/pogreb
// https://youtu.be/CFPcxRN0xp8
// ToDo Заменю после сдачи 2-го спринта
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
func (f *FileStorage) IsExist(id string) bool {
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
func (f *FileStorage) Store(id string, link string) error {
	f.mx.Lock()
	defer f.mx.Unlock()

	a := Alias{Key: id, URL: link}
	err := f.storageWriter.Write(&a)
	if err != nil {
		return err
	}

	return nil
}

// Restore - находит по ID ссылку во внешнем файле, где данные хранятся в формате JSON
func (f *FileStorage) Restore(id string) (link string, err error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	_, err = f.storageReader.file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	for {
		alias, err := f.storageReader.Read()
		if err != nil {
			return "", err
		}

		if alias.Key == id {
			return alias.URL, nil
		}
	}
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
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
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
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *reader) Read() (*Alias, error) {
	event := &Alias{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *reader) Close() error {
	return c.file.Close()
}

// Alias - структура хранения ID и URL во внешнем файле
type Alias struct {
	Key string
	URL string
}
