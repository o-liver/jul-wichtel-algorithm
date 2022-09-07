package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("I'll be your Wichtel!")
	token := os.Getenv("ACCESS_TOKEN")
	fmt.Println("This is your secret:", token)
}