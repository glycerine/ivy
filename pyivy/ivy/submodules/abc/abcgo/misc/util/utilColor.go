package util

import "fmt"

// Abc_ColorTest prints the portable ANSI color/attribute diagnostic from ABC.
func Abc_ColorTest() {
	fmt.Print("Background       ")
	for x := 0; x < 8; x++ {
		fmt.Printf("  [1;4%dm", x)
	}
	fmt.Println()
	for y := 0; y < 2; y++ {
		fmt.Printf("Foreground [%dm   ", y)
		for x := 0; x < 8; x++ {
			fmt.Printf("  \033[%d;3%dm\033[%dm  Hi  \033[0m", y&1, y>>1, x)
		}
		fmt.Println()
	}
	for y := 0; y < 16; y++ {
		fmt.Printf("Foreground [%d;3%dm", y&1, y>>1)
		for x := 0; x < 8; x++ {
			fmt.Printf("  \033[%d;3%dm\033[1;4%dm  Hi  \033[0m", y&1, y>>1, x)
		}
		fmt.Println()
	}
	fmt.Print("\033[4mUnderlined\033[0m\n\033[5mBlinking  \033[0m\n\033[7mInverted  \033[0m\n\033[8mConcealed \033[0m\n")
}
