package main

import (
	"bytes"
	"fmt"
	"image/color"
)

func printWaybar(power bool, level int, mode int) {
	buffer := bytes.Buffer{}

	// color
	if power {
		buffer.WriteString("<span color='")
		outputColor := color.RGBA{G: 170}
		if mode == 1 {
			outputColor = color.RGBA{R: 100, G: 100, B: 250}
		}
		buffer.WriteString(hexColor(outputColor))
		buffer.WriteString("'>")
	}

	// text
	buffer.WriteString("𖣘")
	if power {
		// fmt.Fprint(&buffer, "%{F#00ff80}%{O-3}")
		// buffer.WriteString("<sup>")
		fmt.Fprintf(&buffer, "%s", uInt32ToCircledNumberStr(uint32(level)))
		// buffer.WriteString("</sup>")
		// fmt.Fprint(&buffer, "%{F-}%{O}")
	}

	// end color
	if power {
		buffer.WriteString("</span>")
	}

	fmt.Println(buffer.String())
}
