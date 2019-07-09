package util

import "github.com/fatih/color"

type ColorFunc func(format string, a ...interface{})

var colorsByHostname = map[string]ColorFunc{}
var possibleColors = []func(string, ...interface{}){
	color.Blue,
	color.Red,
	color.Green,
	color.Cyan,
	color.Magenta,
	color.Yellow,
}

func getColor() ColorFunc {
	return possibleColors[len(colorsByHostname)%len(possibleColors)]
}

func GetColorForHostname(hostname string) ColorFunc {
	if f, exists := colorsByHostname[hostname]; exists {
		return f
	}
	colorFunc := getColor()
	colorsByHostname[hostname] = colorFunc
	return colorFunc
}
