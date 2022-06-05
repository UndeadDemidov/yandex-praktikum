package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkStorage_Restore(t *testing.T) {
	type fields struct {
		storage map[string]map[string]string
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
			fields: fields{storage: map[string]map[string]string{
				"xxxx": {
					"1111": "https://ya.ru",
					"2222": "https://yandex.ru",
					"3333": "https://practicum.yandex.ru/",
				},
			}},
			args:     args{"1111"},
			wantLink: "https://ya.ru",
			wantErr:  assert.NoError,
		},
		{
			name: "invalid id",
			fields: fields{storage: map[string]map[string]string{
				"xxxx": {
					"1111": "https://ya.ru",
					"2222": "https://yandex.ru",
					"3333": "https://practicum.yandex.ru/",
				},
			}},
			args:     args{"4444"},
			wantLink: "",
			wantErr:  assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := Storage{
				storage: tt.fields.storage,
			}
			gotLink, err := ls.Restore(context.Background(), tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("Restore(%v)", tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.wantLink, gotLink, "Restore(%v)", tt.args.id)
		})
	}
}

func TestLinkStorage_Store(t *testing.T) {
	type fields struct {
		storage map[string]map[string]string
	}
	type args struct {
		user string
		link string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "store new link 1",
			fields: fields{storage: map[string]map[string]string{
				"xxxx": {
					"1111": "https://ya.ru",
					"2222": "https://yandex.ru",
				},
			}},
			args:    args{"xxxx", "https://practicum.yandex.ru/"},
			wantErr: assert.NoError,
		},
		{
			name:    "store new link 2",
			fields:  fields{storage: map[string]map[string]string{}},
			args:    args{"yyyy", "https://practicum.yandex.ru/"},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				storage: tt.fields.storage,
			}
			ctx := context.Background()
			_, err := ms.Store(ctx, tt.args.user, tt.args.link)
			if !tt.wantErr(t, err, fmt.Sprintf("Store(%v, %v, %v)", ctx, tt.args.user, tt.args.link)) {
				return
			}
			// надо мокать генератор уникальных id
			// assert.Equalf(t, tt.wantId, gotId, "Store(%v, %v, %v)", ctx, tt.args.user, tt.args.link)
		})
	}
}

func TestMemoryStorage_isExist(t *testing.T) {
	type fields struct {
		storage map[string]map[string]string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "check new id",
			fields: fields{storage: map[string]map[string]string{
				"xxxx": {
					"1111": "https://ya.ru",
					"2222": "https://yandex.ru",
				},
			}},
			args: args{"xxxx"},
			want: false,
		},
		{
			name: "check existing id",
			fields: fields{storage: map[string]map[string]string{
				"xxxx": {
					"1111": "https://ya.ru",
					"2222": "https://yandex.ru",
				},
			}},
			args: args{"1111"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Storage{
				storage: tt.fields.storage,
			}
			ctx := context.Background()
			assert.Equalf(t, tt.want, ms.isExist(ctx, tt.args.id), "IsExist(%v, %v)", ctx, tt.args.id)
		})
	}
}
