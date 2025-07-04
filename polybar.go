package main

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"strings"
)

func hexColor(c color.Color) string {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("#%.2x%.2x%.2x", rgba.R, rgba.G, rgba.B)
}

func uInt32ToCircledNumberStr(number uint32) string {
	cache := map[string]string{
		"0": "🄌",
		"1": "➊",
		"2": "➋",
		"3": "➌",
		"4": "➍",
		"5": "➎",
		"6": "➏",
		"7": "➐",
		"8": "➑",
		"9": "➒",
	}

	result := strings.Builder{}
	for _, ch := range fmt.Sprintf("%d", number) {
		if newch, ok := cache[string(ch)]; ok {
			result.WriteString(newch)
		} else {
			result.WriteString(string(ch))
		}
	}
	return result.String()
}

func buildCmd(click string, args ...string) string {
	b := strings.Builder{}

	exec, _ := os.Executable()
	b.WriteString(exec)
	for _, arg := range args {
		b.WriteString(" ")
		b.WriteString(arg)
	}

	return "%{A" + click + ":" + strings.Replace(b.String(), ":", "\\:", -1) + ":}"
}

func printPolybar(power bool, level int) {
	buffer := bytes.Buffer{}

	// color
	if power {
		buffer.WriteString("%{F")
		buffer.WriteString(hexColor(color.RGBA{G: 170}))
		buffer.WriteString("}")
	}

	// left click 1
	// buffer.WriteString(buildCmd("1", "--horizontal-swing"))
	// middle click 2
	// buffer.WriteString(buildCmd("2", "--update"))
	// right click 3
	// buffer.WriteString(buildCmd("3", "--toogle"))
	// scroll up 4
	// buffer.WriteString(buildCmd("4", "--level-up"))
	// scroll down 5
	// buffer.WriteString(buildCmd("5", "--level-down"))

	// text
	buffer.WriteString("")
	if power {
		fmt.Fprint(&buffer, "%{F#00ff80}%{O-3}")
		fmt.Fprintf(&buffer, "%s", uInt32ToCircledNumberStr(uint32(level)))
		fmt.Fprint(&buffer, "%{F-}%{O}")
	}

	// end cmds
	// buffer.WriteString("%{A}")
	// buffer.WriteString("%{A}")
	// buffer.WriteString("%{A}")
	// buffer.WriteString("%{A}")
	// buffer.WriteString("%{A}")

	// end color
	if power {
		buffer.WriteString("%{F-}")
	}

	fmt.Println(buffer.String())
}
