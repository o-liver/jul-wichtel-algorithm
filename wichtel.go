package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	babyBoomersCommaSeperatedList := os.Getenv("BABY_BOOMERS_EMAIL")
	babyBoomers := strings.Split(babyBoomersCommaSeperatedList, ", ")

	millennialsCommaSeperatedList := os.Getenv("MILLENNIALS_EMAIL")
	millennials := strings.Split(millennialsCommaSeperatedList, ", ")

	participants := append(babyBoomers, millennials...)

Start:
	fmt.Println("I'll be your Wichtel!")
	fmt.Println("List of participants this christmas:", participants)
	fmt.Println("Adding millennial's twice into the hat.")
	theHat := append(millennials, millennials...)

	// Shuffle the participants in the hat
	shuffleTheHat(theHat)

	wichtelMatches := make(map[string][]string, len(participants))
	for _, boomer := range babyBoomers {
		var firstMatch, secondMatch string
		firstSlipOfPaperIndex := 0
		firstMatch = theHat[firstSlipOfPaperIndex]
		theHat = removeSlipOfPaperWithIndex(theHat, firstSlipOfPaperIndex)
		fmt.Printf("Found first match for '%v'.\n", boomer)
		for index, slipOfPaper := range theHat { // Iterate over slips of paper in the hat starting at the second entry
			secondMatch = slipOfPaper
			if secondMatch != firstMatch { // If the first and second match are not the same we have the result we want
				theHat = removeSlipOfPaperWithIndex(theHat, index) // overwrite hat removing the first entry
				fmt.Printf("Found second match for '%v', that is not equal to the first match.\n", boomer)
				break // we can break our of the for loop
			} // else we go to the next slip of paper
		}
		wichtelMatches[boomer] = []string{firstMatch, secondMatch}
	}

	fmt.Println("Adding baby boomers twice into the hat.")
	theHat = append(theHat, babyBoomers...)
	theHat = append(theHat, babyBoomers...)

	shuffleTheHat(theHat)

	// Find matches for the millennials
	for _, millennial := range millennials {
		var firstMatch, secondMatch string
		// Make sure the participant did not pull their own name
		for index, slipOfPaper := range theHat {
			firstMatch = slipOfPaper
			if firstMatch != millennial {
				theHat = removeSlipOfPaperWithIndex(theHat, index)
				fmt.Printf("Found first match for '%v'.\n", millennial)
				break
			}
		}

		for index, slipOfPaper := range theHat { // Iterate over slips of paper in the hat starting at the second entry
			secondMatch = slipOfPaper
			// Make sure the participant did not pull their own name
			if secondMatch != millennial && secondMatch != firstMatch {
				theHat = removeSlipOfPaperWithIndex(theHat, index) // overwrite hat removing the first entry
				fmt.Printf("Found second match for '%v', that is not equal to the first match.\n", millennial)
				break // we can break our of the for loop
			}
		}
		wichtelMatches[millennial] = []string{firstMatch, secondMatch}
	}

	if len(theHat) > 0 {
		fmt.Println("Oh-Oh, the last participant found their own name in the hat, starting over...\n")
		goto Start
	}

	// TODO Send out emails
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
