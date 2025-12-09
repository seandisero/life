package life

import (
	"fmt"
	"math/rand/v2"
)

type Life struct {
	grid [][]Soul
}

func life_create() *Life {
	l := Life{}
	l.grid = make([][]Soul, SIZE_Y)
	for i := range SIZE_Y {
		l.grid[i] = make([]Soul, SIZE_X)
	}

	return &l
}

func (l *Life) write_to_world(w *World) error {
	for y := range SIZE_Y {
		for x := range SIZE_X {
			w.set_rune(y/2, y%2, x, l.grid[y][x].Alive)
		}
	}
	return nil
}

func (l *Life) life_fill_random(threshold float32) error {
	if threshold < 0.0 || 1.0 < threshold {
		return fmt.Errorf("malformed threshold")
	}
	for y := range SIZE_Y {
		for x := range SIZE_X {
			r := rand.Float32()
			if r > threshold {
				l.grid[y][x].Alive = false
				continue
			}
			l.grid[y][x].Alive = true
		}
	}
	return nil
}

func (l *Life) add_shape(x, y int, s shape, direction int) {
	for i := 0; i < s.height; i++ {
		for j := 0; j < s.width; j++ {
			scaled_x := j
			if direction == 1 {
				scaled_x = (s.width - 1) - j
			}
			l.grid[(i+y)%SIZE_Y][(j+x)%SIZE_X].Alive = s.shape[i*s.width+scaled_x] == 1
		}
	}
}
