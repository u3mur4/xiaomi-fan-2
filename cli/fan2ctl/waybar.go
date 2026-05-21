package main

import (
	"bytes"
	"fmt"
	"image/color"
)

func printWaybar(online bool, power bool, level int, mode int) {
	buffer := bytes.Buffer{}

	if !online {
		buffer.WriteString("<span color='#888888'>𖣘</span>")
		fmt.Println(buffer.String())
		return
	}

	if power {
		buffer.WriteString("<span color='")
		outputColor := color.RGBA{G: 170}
		if mode == 1 {
			outputColor = color.RGBA{R: 100, G: 100, B: 250}
		}
		buffer.WriteString(hexColor(outputColor))
		buffer.WriteString("'>")
	}

	buffer.WriteString("𖣘")
	if power {
		fmt.Fprintf(&buffer, "%s", uInt32ToCircledNumberStr(uint32(level)))
	}

	if power {
		buffer.WriteString("</span>")
	}

	fmt.Println(buffer.String())
}
