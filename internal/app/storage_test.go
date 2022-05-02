package app

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkStorage_Restore(t *testing.T) {
	type fields struct {
		storage map[string]string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantLink string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "valid id",
			fields: fields{storage: map[string]string{
				"1111": "https://ya.ru",
				"2222": "https://yandex.ru",
				"3333": "https://practicum.yandex.ru/",
			}},
			args:     args{"1111"},
			wantLink: "https://ya.ru",
			wantErr:  assert.NoError,
		},
		{
			name: "invalid id",
			fields: fields{storage: map[string]string{
				"1111": "https://ya.ru",
				"2222": "https://yandex.ru",
				"3333": "https://practicum.yandex.ru/",
			}},
			args:     args{"4444"},
			wantLink: "",
			wantErr:  assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := LinkStorage{
				storage: tt.fields.storage,
			}
			gotLink, err := ls.Restore(tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("Restore(%v)", tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.wantLink, gotLink, "Restore(%v)", tt.args.id)
		})
	}
}

func TestLinkStorage_Store(t *testing.T) {
	type fields struct {
		storage map[string]string
	}
	type args struct {
		id   string
		link string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "store new link",
			fields: fields{storage: map[string]string{
				"1111": "https://ya.ru",
				"2222": "https://yandex.ru",
			}},
			args:    args{"3333", "https://practicum.yandex.ru/"},
			wantErr: assert.NoError,
		},
		{
			//ToDo - с реализацией работы с дубликатами переписать тест
			name: "store existing link",
			fields: fields{storage: map[string]string{
				"1111": "https://ya.ru",
				"2222": "https://yandex.ru",
			}},
			args:    args{"3333", "https://yandex.ru"},
			wantErr: assert.NoError,
		},
		{
			name: "store with same id",
			fields: fields{storage: map[string]string{
				"1111": "https://ya.ru",
				"2222": "https://yandex.ru",
			}},
			args:    args{"2222", "https://practicum.yandex.ru/"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := LinkStorage{
				storage: tt.fields.storage,
			}
			err := ls.Store(tt.args.id, tt.args.link)
			if !tt.wantErr(t, err, fmt.Sprintf("Store(%v)", tt.args.link)) {
				return
			}
		})
	}
}
