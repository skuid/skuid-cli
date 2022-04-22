package main

import "github.com/gookit/color"

func Println(args ...interface{}) {
	color.Blue.Println(args...)
}

func Printf(formatString string, args ...interface{}) {
	color.Red.Printf(formatString, args...)
}

func init() {
	color.Enable = true
}
