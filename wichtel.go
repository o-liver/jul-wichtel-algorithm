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
	babyBoomers := strings.Split(babyBoomersCommaSeperatedList, ", ")

	millennialsCommaSeperatedList := os.Getenv("MILLENNIALS_EMAIL")
	millennials := strings.Split(millennialsCommaSeperatedList, ", ")

	participants := append(babyBoomers, millennials...)
	fmt.Println("List of participants this christmas:", participants)

	fmt.Println("Adding millennial's twice into the hat")
	theHat := append(millennials, millennials...)

	// Shuffle the participants in the hat
	shuffleTheHat(theHat)

	wichtelMatches := make(map[string][]string, len(participants))
	for _, boomer := range babyBoomers {
		var firstMatch, secondMatch string
		firstSlipOfPaperIndex := 0
		firstMatch = theHat[firstSlipOfPaperIndex]
		theHat = removeSlipOfPaperWithIndex(theHat, firstSlipOfPaperIndex)
		for index, slipOfPaper := range theHat { // Iterate over slips of paper in the hat starting at the second entry
			secondMatch = slipOfPaper
			if secondMatch != firstMatch { // If the first and second match are not the same we have the result we want
				theHat = removeSlipOfPaperWithIndex(theHat, index) // overwrite hat removing the first entry
				break // we can break our of the for loop
			} // else we go to the next slip of paper
		}
		wichtelMatches[boomer] = []string{firstMatch, secondMatch}
	}

	fmt.Println("Boomer matches:", wichtelMatches) // TODO: Remove this debug message!
	fmt.Println("The hat afterwards:", theHat) // TODO: Remove this debug message!

}

func removeSlipOfPaperWithIndex(hat []string, index int) []string {
	newHat := make([]string, 0)
	newHat = append(newHat, hat[:index]...)   // Retrieve entries before index
	newHat = append(newHat, hat[index+1:]...) // Retrieve entries after index
	return newHat
}

func shuffleTheHat(hat []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(hat), func(i, j int) {
		hat[i], hat[j] = hat[j], hat[i]
	})
}

func sendOutEmailsAndDeleteThemAfter(wichtelMatches map[string][]string) {
	token := os.Getenv("ACCESS_TOKEN")
	fmt.Println("This is your secret:", token)
}
