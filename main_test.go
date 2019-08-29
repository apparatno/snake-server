package main

import (
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
