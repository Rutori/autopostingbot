package utility

import (
	"log"

	"github.com/fatih/color"
)

// GreenLog is a log.Println with green in it
func GreenLog(s string) {
	log.Println(color.GreenString(s))
}

// YellowLog is a log.Println with green in it
func YellowLog(s string) {
	log.Println(color.YellowString(s))
}

// BlueLog is a log.Println with blue in it
func BlueLog(s string) {
	log.Println(color.BlueString(s))
}
