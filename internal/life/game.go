package life

import (
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

type LifeGame struct {
	conf          *config
	life          *Life
	evolution     *Life
	world         *World
	world_size_x  int
	world_size_y  int
	window_width  int
	window_height int
	wait_group    sync.WaitGroup
}

var SIZE_X = 0
var SIZE_Y = 0

func Init_life_game() (*LifeGame, error) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}

	width := w - 2
	hh := h * 2
	height := hh - 2

	SIZE_X = width - (width % 8)
	SIZE_Y = height - (height % 16)

	lg := LifeGame{}
	lg.life = life_create()
	lg.evolution = life_create()
	lg.world = world_make()
	lg.window_width = width
	lg.window_height = height

	lg.wait_group = sync.WaitGroup{}

	conf, err := lg.select_config()
	if err != nil {
		return nil, err
	}

	lg.conf = conf
	lg.add_life_from_config()
	return &lg, nil
}

func (lg *LifeGame) Run() {
	const speed = 60

	lg.wait_group.Add(1)
	go lg.print()
	time.Sleep(speed * time.Millisecond)
	lg.wait_group.Wait()
	for {
		lg.live()
		lg.wait_group.Add(1)
		go lg.print()
		time.Sleep(speed * time.Millisecond)
		lg.wait_group.Wait()
	}
}

func (lg *LifeGame) add_life_from_config() {
	switch lg.conf.pat {
	case RANDOM:
		lg.life.life_fill_random(0.5)
	case GUN:
		rd := rand.Int()
		lg.life.add_shape(16, 16, gosper_glider_gun, rd%2)
	case SHIPS:
		shapes := []shape{lwss, hwss, copperhead, weekender, butterfly, plate_v1}
		buffer := 8
		next := 0
		for {
			r := rand.Int()
			rd := rand.Int()
			shape := shapes[r%len(shapes)]
			if next+shape.height+buffer > SIZE_Y {
				break
			}
			lg.life.add_shape(r%SIZE_X, next+buffer, shape, rd%2)
			next += shape.height + buffer
		}
	}

}

func print_options(op []string) {
	w := 32
	h := len(op) + 2
	skip_len := 0
	for i := range h {
		for j := range w {
			if i == 0 {
				if j == 0 {
					fmt.Print("╔")
					continue
				}
				if j == w-1 {
					fmt.Print("╗")
					continue
				}
				fmt.Print("═")
				continue
			} else if i == h-1 {
				if j == 0 {
					fmt.Print("╚")
					continue
				}
				if j == w-1 {
					fmt.Print("╝")
					continue
				}
				fmt.Print("═")
				continue
			}
			if skip_len == 0 {
				fmt.Printf("║ %s", op[i-1])
				skip_len = len(op[i-1]) + 2
			} else if j < skip_len {
				continue
			} else if j == w-2 {
				fmt.Printf("║")
			}
			fmt.Print(" ")
		}
		fmt.Print("\n")
		skip_len = 0
	}
}

func (lg *LifeGame) select_config() (*config, error) {
	op := []string{"0. random", "1. gun", "2. ships"}
	print_options(op)
	var pat pattern
	_, err := fmt.Scan(&pat)
	if err != nil {
		return nil, err
	}
	if pat < 0 || pat > 2 {
		return nil, fmt.Errorf("invalid selection")
	}
	conf := config{}
	conf.pat = pat
	fmt.Print("\033[2J")
	return &conf, nil
}

func (lg *LifeGame) print() {
	fmt.Print("\033[H")
	fmt.Print("\033[?25l")
	defer lg.wait_group.Done()
	lg.world.print_world(lg.window_width, lg.window_height)
}

func (lg *LifeGame) evolve(x, y int) {
	is_alive := lg.life.grid[y][x].Alive
	alive_neightbor_count := 0
	lg.evolution.grid[y][x].Alive = false
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
			if lg.life.grid[yy][xx].Alive {
				alive_neightbor_count++
			}
		}
	}
	if !is_alive && alive_neightbor_count == 3 {
		lg.evolution.grid[y][x].Alive = true
		return
	}
	if is_alive && (alive_neightbor_count == 2 || alive_neightbor_count == 3) {
		lg.evolution.grid[y][x].Alive = true
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
