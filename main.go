package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

const NULL = ' '
const UPPER = '\U00002580'
const LOWER = '\U00002584'
const FULL = '\U00002588'

var SIZE_X = 512 + 256 + 64
var SIZE_Y = 512 + 32

type soul struct {
	alive bool
}

type life struct {
	grid [][]soul
}

type world struct {
	grid [][]rune
	buff *bufio.Writer
}

type LifeGame struct {
	life          *life
	evolution     *life
	world         *world
	window_width  int
	window_height int
	wait_group    sync.WaitGroup
}

func init_life_game(width, height int) *LifeGame {
	lg := LifeGame{}
	lg.life = life_create()
	lg.evolution = life_create()
	lg.world = world_make()
	lg.window_width = width
	lg.window_height = height

	lg.wait_group = sync.WaitGroup{}
	return &lg
}

func (l *life) write_to_world(w *world) error {
	for y := range SIZE_Y {
		for x := range SIZE_X {
			w.set_rune(y/2, y%2, x, l.grid[y][x].alive)
		}
	}
	return nil
}

func life_create() *life {
	l := life{}
	l.grid = make([][]soul, SIZE_Y)
	for i := range SIZE_Y {
		l.grid[i] = make([]soul, SIZE_X)
	}

	return &l
}

func (l *life) life_fill_random() error {
	for y := range SIZE_Y {
		for x := range SIZE_X {
			r := rand.Float32()
			if r > 0.3 {
				l.grid[y][x].alive = false
				continue
			}
			l.grid[y][x].alive = true
		}
	}
	return nil
}

func world_make() *world {
	w := world{}
	w.grid = make([][]rune, SIZE_Y/2)
	for i := range SIZE_Y / 2 {
		w.grid[i] = make([]rune, SIZE_X)
	}
	for y := range SIZE_Y / 2 {
		for x := range SIZE_X {
			w.grid[y][x] = ' '
		}
	}

	w.buff = bufio.NewWriterSize(os.Stdout, SIZE_X*SIZE_Y+SIZE_Y+3)

	return &w
}

func (w *world) set_rune(x, sx, y int, alive bool) {
	if sx == 0 {
		if alive {
			switch w.grid[x][y] {
			case UPPER:
				break
			case FULL:
				break
			case LOWER:
				w.grid[x][y] = FULL
			case NULL:
				w.grid[x][y] = UPPER
			}
		} else {
			switch w.grid[x][y] {
			case UPPER:
				w.grid[x][y] = NULL
			case FULL:
				w.grid[x][y] = LOWER
			case LOWER:
				break
			case NULL:
				break
			}

		}
	}
	if sx == 1 {
		if alive {
			switch w.grid[x][y] {
			case LOWER:
				break
			case FULL:
				break
			case UPPER:
				w.grid[x][y] = FULL
			case NULL:
				w.grid[x][y] = LOWER
			}
		} else {
			switch w.grid[x][y] {
			case LOWER:
				w.grid[x][y] = NULL
			case FULL:
				w.grid[x][y] = UPPER
			case UPPER:
				break
			case NULL:
				break
			}

		}
	}
}

func (w *world) print_world(width, height int) {
	w.buff.Reset(w.buff)
	for range (height - SIZE_Y) / 4 {
		for range width {
			w.buff.WriteString(" ")
		}
		w.buff.WriteString("\n")
	}
	for range (width - SIZE_X) / 2 {
		w.buff.WriteString(" ")
	}
	w.buff.WriteString("╔")
	for range SIZE_X {
		w.buff.WriteString("═")
	}
	w.buff.WriteString("╗")
	w.buff.WriteString("\n")
	for y := range SIZE_Y/2 - 1 {
		for range (width - SIZE_X) / 2 {
			w.buff.WriteString(" ")
		}
		w.buff.WriteString("║")
		w.buff.WriteString(string(w.grid[y]))
		w.buff.WriteString("║")
		w.buff.WriteString("\n")
	}
	for range (width - SIZE_X) / 2 {
		w.buff.WriteString(" ")
	}
	w.buff.WriteString("╚")
	for range SIZE_X {
		w.buff.WriteString("═")
	}
	w.buff.WriteString("╝")
	w.buff.Flush()
}

func (lg *LifeGame) print() {
	fmt.Print("\033[H")
	fmt.Print("\033[?25l")
	defer lg.wait_group.Done()
	lg.world.print_world(lg.window_width, lg.window_height)
}

func (lg *LifeGame) evolve(x, y int) {
	is_alive := lg.life.grid[y][x].alive
	alive_neightbor_count := 0
	lg.evolution.grid[y][x].alive = false
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			yy := y + i
			if yy == -1 {
				yy = SIZE_Y - 1
			}
			if yy == SIZE_Y {
				yy = 0
			}
			xx := x + j
			if xx == -1 {
				xx = SIZE_X - 1
			}
			if xx == SIZE_X {
				xx = 0
			}
			if lg.life.grid[yy][xx].alive {
				alive_neightbor_count++
			}
		}
	}
	if !is_alive && alive_neightbor_count == 3 {
		lg.evolution.grid[y][x].alive = true
		return
	}
	if is_alive && (alive_neightbor_count == 2 || alive_neightbor_count == 3) {
		lg.evolution.grid[y][x].alive = true
	}
}

func (lg *LifeGame) evolve_chunk(x, y, chunk_size int) {
	defer lg.wait_group.Done()
	for i := y; i < y+chunk_size; i++ {
		for j := x; j < x+chunk_size; j++ {
			lg.evolve(j, i)
		}
	}
}

func (lg *LifeGame) live() {
	chunk_size := 8
	for y := 0; y < SIZE_Y; y = y + chunk_size {
		for x := 0; x < SIZE_X; x = x + chunk_size {
			lg.wait_group.Add(1)
			go lg.evolve_chunk(x, y, chunk_size)
		}
	}

	lg.wait_group.Wait()

	for y := 0; y < SIZE_Y; y++ {
		copy(lg.life.grid[y], lg.evolution.grid[y])
	}

	lg.life.write_to_world(lg.world)
}

func (lg *LifeGame) run() {
	lg.life.life_fill_random()
	lg.life.write_to_world(lg.world)
	for {
		lg.live()
		lg.wait_group.Add(1)
		go lg.print()
		time.Sleep(100 * time.Millisecond)
		lg.wait_group.Wait()
	}
}

func main() {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}

	width := w - 2
	hh := h * 2
	height := hh - 2

	SIZE_X = width - (width % 8)
	SIZE_Y = height - (height % 16)

	life_game := init_life_game(w, h*2)
	life_game.run()
}
