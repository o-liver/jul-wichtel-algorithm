package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const subjectID = "jul-wichtel-algorithm"

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

	gmailService := setupGmailService()

	sendOutEmails(wichtelMatches, gmailService)

	deleteWichtelEmails(gmailService)
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromEnvironment("TOKEN_JSON")
	if err != nil {
		fmt.Errorf("something went wrong reading the authentication json: %v", err)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a environment.
func tokenFromEnvironment(envVarName string) (*oauth2.Token, error) {
	tokenJson := os.Getenv(envVarName)
	tok := &oauth2.Token{}
	err := json.NewDecoder(strings.NewReader(tokenJson)).Decode(tok)
	return tok, err
}

func setupGmailService() *gmail.Service {
	ctx := context.Background()

	credentialsJSON := os.Getenv("CREDENTIALS_JSON")

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(credentialsJSON), gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	return gmailService
}

func sendOutEmails(wichtelMatches map[string][]string, gmailService *gmail.Service) {
	for wichtel, gifted := range wichtelMatches {
		err := sendEmail(
			wichtel,
			fmt.Sprintf("Congratulations, you are secret santa to: %s and %s", gifted[0], gifted[1]),
			gmailService)
		if err != nil {
			log.Fatal("Could not send email: ", err)
		}
		fmt.Printf("Sent out email to '%v'\n", wichtel)
	}

}

func sendEmail(recipient, body string, gmailService *gmail.Service) error {
	var message gmail.Message

	email := []byte("From: 'me'\r\n" +
		"To:  " + recipient + "\r\n" +
		"Subject: Your wichtel result from the " + subjectID + "\r\n" +
		"\r\n" + body)

	message.Raw = base64.StdEncoding.EncodeToString(email)
	message.Raw = strings.Replace(message.Raw, "/", "_", -1)
	message.Raw = strings.Replace(message.Raw, "+", "-", -1)
	message.Raw = strings.Replace(message.Raw, "=", "", -1)

	_, err := gmailService.Users.Messages.Send("me", &message).Do()
	return err
}

func deleteWichtelEmails(gmailService *gmail.Service) {
	response, err := gmailService.Users.Messages.List("me").Q(subjectID).Do()
	if err != nil {
		log.Fatal("Could not read list of wichtel mails: ", err)
	}
	fmt.Printf("Found %v messages with subject '%s'\n", len(response.Messages), subjectID)
	for index, message := range response.Messages {
		err = gmailService.Users.Messages.Delete("me", message.Id).Do()
		if err != nil {
			log.Fatalf("Could not delete message with ID '%v'; err: %v", message.Id, err)
		}
		fmt.Printf("Deleted messages number %v\n", index + 1)
	}
}
