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
		name     string
		args     args
		wantLink string
		wantErr  assert.ErrorAssertionFunc
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
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
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
