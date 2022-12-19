package cache

import (
	"testing"
)

func Test_calcIndex(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "0",
			args: args{
				size: 1 << 0,
			},
			want: 0,
		},
		{
			name: "1",
			args: args{
				size: 1 << 1,
			},
			want: 1,
		},
		{
			name: "2",
			args: args{
				size: 1 << 2,
			},
			want: 2,
		},
		{
			name: "24",
			args: args{
				size: 1 << 24,
			},
			want: 24,
		},
		{
			name: "25",
			args: args{
				size: 1<<24 - 1,
			},
			want: 24,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcIndex(tt.args.size); got != tt.want {
				t.Errorf("calcIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_log2(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "0",
			args: args{
				size: 1 << 0,
			},
			want: 0,
		},
		{
			name: "1",
			args: args{
				size: 1 << 1,
			},
			want: 1,
		},
		{
			name: "2",
			args: args{
				size: 1 << 2,
			},
			want: 2,
		},
		{
			name: "24",
			args: args{
				size: 1 << 24,
			},
			want: 24,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := log2(tt.args.size); got != tt.want {
				t.Errorf("log2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache(t *testing.T) {
	buf := New(0)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(0)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1 << 0)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1 << 0)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1 << 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1 << 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1 << 2)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1 << 2)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1 << 24)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1 << 24)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1<<24 - 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1<<24 - 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)

	buf = New(1<<24 + 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
	Delete(buf)
	buf = New(1<<24 + 1)
	t.Logf("size: %d", cap(buf))
	t.Logf("address: %p", buf)
}

func TestMake(t *testing.T) {
	buf := make([]byte, 0)
	t.Logf("size: %d", cap(buf))
	buf = make([]byte, 0, 2)
	t.Logf("size: %d", cap(buf))
}
