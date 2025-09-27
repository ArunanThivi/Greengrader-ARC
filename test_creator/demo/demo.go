package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Print("Enter input: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := scanner.Text()
		fmt.Println("You entered:", input)
	}
}
