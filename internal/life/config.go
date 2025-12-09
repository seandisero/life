package life

type pattern int

const (
	RANDOM = iota
	GUN
	SHIPS
)

type config struct {
	size_x int
	size_y int
	pat    pattern
}
