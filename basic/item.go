package basic

type ListType uint8

const (
	Probation ListType = iota
	Protection
	Window
)

type Item struct {
	Belong ListType
	Key    string
	Val    string
}
