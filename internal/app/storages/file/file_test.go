package file

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_isExist(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "searching first item",
			args: args{id: "1111"},
			want: true,
		},
		{
			name: "searching last item",
			args: args{id: "4444"},
			want: true,
		},
		{
			name: "searching non existing item",
			args: args{id: "5555"},
			want: false,
		},
	}

	filename := "file_storage.json"
	fs, err := NewStorage(filename)
	require.NoError(t, err)
	defer func(fs *Storage) {
		err := fs.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(fs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, fs.isExist(context.Background(), tt.args.id), "IsExist(%v)", tt.args.id)
		})
	}
}

func TestFileStorage_Restore(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		wantErr  assert.ErrorAssertionFunc
		name     string
		wantLink string
		args     args
	}{
		{
			name:     "searching first item",
			args:     args{id: "1111"},
			wantLink: "https://ya.ru",
			wantErr:  assert.NoError,
		},
		{
			name:     "searching last item",
			args:     args{id: "4444"},
			wantLink: "https://github.com/spf13/afero",
			wantErr:  assert.NoError,
		},
		{
			name:     "searching non existing item",
			args:     args{id: "5555"},
			wantLink: "",
			wantErr:  assert.Error,
		},
	}

	filename := "file_storage.json"
	fs, err := NewStorage(filename)
	require.NoError(t, err)
	defer func(fs *Storage) {
		err := fs.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(fs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, err := fs.Restore(context.Background(), tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("Restore(%v)", tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.wantLink, gotLink, "Restore(%v)", tt.args.id)
		})
	}
}

func TestFileStorage_Store(t *testing.T) {
	type args struct {
		user string
		link string
	}
	tests := []struct {
		wantErr assert.ErrorAssertionFunc
		name    string
		args    args
	}{
		{
			name: "adding new item 1",
			args: args{
				user: "xxxx",
				link: "https://ya.ru",
			},
			wantErr: assert.NoError,
		},
		{
			name: "adding new item 2",
			args: args{
				user: "xxxx",
				link: "https://yandex.ru",
			},
			wantErr: assert.NoError,
		},
	}

	filename := "test_storage.txt"
	fs, err := NewStorage(filename)
	require.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatalln(err)
		}
	}(filename)
	defer func(fs *Storage) {
		err := fs.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(fs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := fs.Store(ctx, tt.args.user, tt.args.link)
			if !tt.wantErr(t, err, fmt.Sprintf("Store(%v, %v, %v)", ctx, tt.args.user, tt.args.link)) {
				return
			}
			// assert.Equalf(t, tt.wantId, gotId, "Store(%v, %v, %v)", tt.args.ctx, tt.args.user, tt.args.link)
		})
	}
}

func TestStorage_GetUserStorage(t *testing.T) {
	type args struct {
		user string
	}
	tests := []struct {
		want map[string]string
		name string
		args args
	}{
		{
			name: "empty bucket",
			args: args{user: "yyyy"},
			want: make(map[string]string),
		},
		{
			name: "full bucket",
			args: args{user: "xxxx"},
			want: map[string]string{"1111": "https://ya.ru", "2222": "https://yandex.ru", "3333": "https://go.dev", "4444": "https://github.com/spf13/afero"},
		},
	}

	filename := "file_storage.json"
	fs, err := NewStorage(filename)
	require.NoError(t, err)
	defer func(fs *Storage) {
		err := fs.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(fs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, fs.GetUserStorage(context.Background(), tt.args.user), "GetUserStorage(context.Background(), %v)", tt.args.user)
		})
	}
}
