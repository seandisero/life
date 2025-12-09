package life

import (
	"bufio"
	"fmt"
	"os"
)

var MAX_AGE = 255
var AGE_INCREMENT = 2

var warm_color_scale = 0.3

var blue = LinearColor{0.447 * 0.5, 0.8549 * 0.5, 0.945 * 0.5}
var red = LinearColor{0.7 * warm_color_scale, 0.2 * warm_color_scale, 0.1 * warm_color_scale}
var orage = LinearColor{0.98823 * warm_color_scale, 0.584313 * warm_color_scale, 0.062745 * warm_color_scale}
var white = LinearColor{0.88 * 0.4, 0.89 * 0.4, 0.9 * 0.4}
var black = LinearColor{0.0, 0.0, 0.0}

type Space struct {
	r         rune
	fg_color  LinearColor
	bg_color  LinearColor
	top_alive bool
	bot_alive bool
	top_age   int
	bot_age   int
}

type World struct {
	grid [][]Space
	buff *bufio.Writer
}

type LinearColor struct {
	r float64
	g float64
	b float64
}

func (lc *LinearColor) normalize() {
	lc.r /= float64(255)
	lc.g /= float64(255)
	lc.b /= float64(255)
}

func (c *LinearColor) set(r, g, b float64) {
	c.r = r
	c.g = g
	c.b = b
}

func (col *LinearColor) lerp(x LinearColor, y LinearColor, t float64) {
	col.r = x.r + (y.r-x.r)*t
	col.g = x.g + (y.g-x.g)*t
	col.b = x.b + (y.b-x.b)*t
}

func (c *LinearColor) age_to_rgb(age int) {
	nage := float64(age) / float64(MAX_AGE)

	t := nage * 4

	if nage < 0.25 {
		c.lerp(blue, white, (t))
	} else if nage < 0.5 {
		c.lerp(white, orage, (t)-1)
	} else if nage < 0.75 {
		c.lerp(orage, red, (t)-2)
	} else {
		c.lerp(red, black, (t)-3)
	}

}

func (s *Space) set_top_alive(alive bool) {
	s.top_alive = alive
	if s.top_alive {
		s.top_age = 0
		if s.bot_alive {
			s.r = FULL
			return
		}
	}
	if !s.top_alive {
		s.top_age += AGE_INCREMENT
		if s.bot_alive {
			s.r = LOWER
		}
		if !s.bot_alive {
			if s.top_age > s.bot_age {
				s.r = UPPER
			} else {
				s.r = LOWER
			}
		}
	}

}

func (s *Space) set_bottom_alive(alive bool) {
	s.bot_alive = alive
	if s.bot_alive {
		s.bot_age = 0
		if s.top_alive {
			s.r = FULL
			return
		}
	}
	if !s.bot_alive {
		s.bot_age += AGE_INCREMENT
		if s.top_alive {
			s.r = UPPER
		}
		if !s.top_alive {
			if s.bot_age > s.top_age {
				s.r = LOWER
			} else {
				s.r = UPPER
			}
		}
	}

}

func (s *Space) format() string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm%c\033[38;2;255;255;255m\033[48;2;0;0;0m",
		int(s.fg_color.r*255), int(s.fg_color.g*255), int(s.fg_color.b*255),
		int(s.bg_color.r*255), int(s.bg_color.g*255), int(s.bg_color.b*255),
		s.r)
}

func (s *Space) to_string() string {
	if s.top_age > MAX_AGE {
		s.top_age = MAX_AGE - 1
	}
	if s.bot_age > MAX_AGE {
		s.bot_age = MAX_AGE - 1
	}

	if s.r == UPPER {
		if s.top_age == 0 {
			s.fg_color.set(1.0, 1.0, 1.0)
		} else {
			s.fg_color.age_to_rgb(s.top_age)
		}
		s.bg_color.age_to_rgb(s.bot_age)
		return s.format()

	} else if s.r == LOWER {
		if s.bot_age == 0 {
			s.fg_color.set(1.0, 1.0, 1.0)
		} else {
			s.fg_color.age_to_rgb(s.bot_age)
		}
		s.bg_color.age_to_rgb(s.top_age)
		return s.format()

	} else if s.r == FULL {
		s.fg_color.set(1.0, 1.0, 1.0)
		s.bg_color.set(1.0, 1.0, 1.0)
		return s.format()

	}
	s.bg_color.age_to_rgb(min(s.bot_age, s.top_age))
	return s.format()
}

func world_make() *World {
	w := World{}
	w.grid = make([][]Space, SIZE_Y/2)
	for i := range SIZE_Y / 2 {
		w.grid[i] = make([]Space, SIZE_X)
	}
	for y := range SIZE_Y / 2 {
		for x := range SIZE_X {
			w.grid[y][x].r = ' '
			w.grid[y][x].top_age = MAX_AGE
			w.grid[y][x].bot_age = MAX_AGE
			w.grid[y][x].bg_color = LinearColor{0.0, 0.0, 0.0}
		}
	}

	w.buff = bufio.NewWriter(os.Stdout)

	return &w
}

func (w *World) set_rune(x, sx, y int, alive bool) {
	if sx == 0 {
		w.grid[x][y].set_top_alive(alive)
	}
	if sx == 1 {
		w.grid[x][y].set_bottom_alive(alive)
	}
}

func (w *World) print_world(width, height int) {
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
		for x := range SIZE_X {
			w.buff.WriteString(w.grid[y][x].to_string())
		}
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
