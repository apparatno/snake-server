package main

import (
	"reflect"
	"testing"
)

func TestCalculateNextPixel(t *testing.T) {
	type arg struct {
		snake     []int
		direction string
	}
	tests := []struct {
		name string
		arg  arg
		want int
	}{
		{
			name: "moves right",
			arg: arg{
				snake:     []int{112, 111, 110},
				direction: "R",
			},
			want: 113,
		},
		{
			name: "moves left",
			arg: arg{
				snake:     []int{110, 111, 112},
				direction: "L",
			},
			want: 109,
		},
		{
			name: "moves up",
			arg: arg{
				snake:     []int{112, 111, 110},
				direction: "U",
			},
			want: 92,
		},
		{
			name: "moves down",
			arg: arg{
				snake:     []int{112, 111, 110},
				direction: "D",
			},
			want: 132,
		},
		{
			name: "appears on the left when going out on the right",
			arg: arg{
				snake:     []int{39, 38, 37},
				direction: "R",
			},
			want: 20,
		},
		{
			name: "appears on the right when going out on the left",
			arg: arg{
				snake:     []int{20, 21, 22},
				direction: "L",
			},
			want: 39,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := calculateNextPixel(tt.arg.snake, tt.arg.direction); res != tt.want {
				t.Errorf("want %d got %d", tt.want, res)
			}

		})
	}
}

func Test_consumedFruit(t *testing.T) {
	consumeAtIndex := 1
	fruits := []fruit{
		fruit{
			position: consumeAtIndex,
		},
	}

	type arg struct {
		snekHeadPosition int
		fruits           []fruit
	}

	tests := []struct {
		name string
		args arg
		want bool
	}{
		{
			name: "consumed fruit if head at fruit",
			want: true,
			args: arg{
				snekHeadPosition: consumeAtIndex,
			},
		},
		{
			name: "did not fruit if head at not fruit",
			want: false,
			args: arg{
				snekHeadPosition: 999,
			},
		},
	}

	for _, tt := range tests {
		if got, _ := consumedFruit(tt.args.snekHeadPosition, fruits); got != tt.want {
			t.Errorf("consumedFruit() = %v, want %v", got, tt.want)
		}
	}
}

func Test_moveMotherfuckingSnake(t *testing.T) {
	type args struct {
		snake     []int
		direction string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "moves left",
			args: args{
				snake:     []int{110, 111, 112},
				direction: "L",
			},
			want: []int{109, 110, 111},
		},
		{
			name: "moves right",
			args: args{
				snake:     []int{112, 111, 110},
				direction: "R",
			},
			want: []int{113, 112, 111},
		},
		{
			name: "going left turns up",
			args: args{
				snake:     []int{110, 111, 112},
				direction: "U",
			},
			want: []int{90, 110, 111},
		},
		{
			name: "going left turns down",
			args: args{
				snake:     []int{110, 111, 112},
				direction: "D",
			},
			want: []int{130, 110, 111},
		},
		{
			name: "going right turns up",
			args: args{
				snake:     []int{112, 111, 110},
				direction: "U",
			},
			want: []int{92, 112, 111},
		},
		{
			name: "going right turns down",
			args: args{
				snake:     []int{112, 111, 110},
				direction: "D",
			},
			want: []int{132, 112, 111},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := moveMotherfuckingSnake(tt.args.snake, tt.args.direction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("moveMotherfuckingSnake() = %v, want %v", got, tt.want)
			}
		})
	}
}
