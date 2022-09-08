package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("I'll be your Wichtel!")

	babyBoomersCommaSeperatedList := os.Getenv("BABY_BOOMERS_EMAIL")
	babyBoomers := strings.Split(babyBoomersCommaSeperatedList, ",")

	millennialsCommaSeperatedList := os.Getenv("MILLENNIALS_EMAIL")
	millennials := strings.Split(millennialsCommaSeperatedList, ",")

	participants := append(babyBoomers, millennials...)
	fmt.Println("List of participants this christmas:", participants)

	fmt.Println("Adding millennial's twice into the hat")
	theHat := append(millennials, millennials...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(theHat), func(i, j int) {
		theHat[i], theHat[j] = theHat[j], theHat[i]
	})
	fmt.Println("After mixing the hat we got:", theHat)



}

func sendOutEmailsAndDeleteThemAfter(wichtelMatches map[string]string) {
	token := os.Getenv("ACCESS_TOKEN")
	fmt.Println("This is your secret:", token)
}