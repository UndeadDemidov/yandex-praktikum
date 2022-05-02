package app

import (
	"testing"
)

func TestIsURL(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "well formatted URL 1",
			args: args{"https://habr.com/ru/post/66931/"},
			want: true,
		},
		{
			name: "well formatted URL 2",
			args: args{"https://habr.com/ru/post/66931/?"},
			want: true,
		},
		{
			name: "well formatted URL 3",
			args: args{"http://habr.com/ru/post/66931/"},
			want: true,
		},
		{
			name: "well formatted URL 4",
			args: args{"https://ya.ru"},
			want: true,
		},
		{
			name: "bad URL 1",
			args: args{"ya"},
			want: false,
		},
		{
			name: "bad URL 2",
			args: args{"1234"},
			want: false,
		},
		{
			name: "bad URL 3",
			args: args{"https://ya.ru/!@#$%^&*()"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.args.str); got != tt.want {
				t.Errorf("IsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
