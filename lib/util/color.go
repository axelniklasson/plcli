package util

import (
	"sync"

	"github.com/fatih/color"
)

type safeColors struct {
	v   map[string]ColorFunc
	mux sync.Mutex
}

// ColorFunc is a wrapper around a fatih/color function that outputs colored text to stdout
type ColorFunc func(format string, a ...interface{})

// var colorsByHostname = map[string]ColorFunc{}
var mutexedColors = safeColors{v: map[string]ColorFunc{}}
var possibleColors = []func(string, ...interface{}){
	color.Blue,
	color.Red,
	color.Green,
	color.Cyan,
	color.Magenta,
	color.Yellow,
}

func getColor() ColorFunc {
	return possibleColors[len(mutexedColors.v)%len(possibleColors)]
}

// GetColorForHostname returns a color func that prints colored output to stdout (always same color for given hostname)
func GetColorForHostname(hostname string) ColorFunc {
	mutexedColors.mux.Lock()
	if f, exists := mutexedColors.v[hostname]; exists {
		mutexedColors.mux.Unlock()
		return f
	}
	colorFunc := getColor()
	mutexedColors.v[hostname] = colorFunc
	mutexedColors.mux.Unlock()
	return colorFunc
}
