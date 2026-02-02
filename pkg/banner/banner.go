package banner

import (
	"fmt"
	"strings"
)

// Rainbow renkleri (ANSI)
var rainbow = []int{31, 33, 32, 36, 34, 35} // R,Y,G,C,B,M

// PrintRainbow ASCII art'ı gökkuşağı renkleriyle yazdırır
func PrintRainbow(ascii string) {
	lines := strings.Split(strings.TrimSpace(ascii), "\n")
	for i, line := range lines {
		for j, r := range line {
			c := rainbow[(i+j)%len(rainbow)]
			fmt.Printf("\033[%dm%c\033[0m", c, r)
		}
		fmt.Println()
	}
}

// ErosHitASCII ErosHit logosu
const ErosHitASCII = `                                                      /$$      
                                                     | $$      
  /$$$$$$   /$$$$$$   /$$$$$$   /$$$$$$$     /$$$$$$$| $$$$$$$ 
 /$$__  $$ /$$__  $$ /$$__  $$ /$$_____/    /$$_____/| $$__  $$
| $$$$$$$$| $$  \__/| $$  \ $$|  $$$$$$    |  $$$$$$ | $$  \ $$
| $$_____/| $$      | $$  | $$ \____  $$    \____  $$| $$  | $$
|  $$$$$$$| $$      |  $$$$$$/ /$$$$$$$//$$ /$$$$$$$/| $$  | $$
 \_______/|__/       \______/ |_______/|__/|_______/ |__/  |__/`
