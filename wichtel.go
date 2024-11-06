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

type slipOfPaper struct {
	email string
	group string
}

func main() {
	numberOfTries := 0

	babyBoomersCommaSeparatedList := os.Getenv("TEST_BABY_BOOMERS_EMAIL")
	fmt.Printf("Baby boomers comma separated list: %v\n", babyBoomersCommaSeparatedList)
	babyBoomers := strings.Split(babyBoomersCommaSeparatedList, ", ")
	fmt.Printf("Baby boomers: %v\n", babyBoomers)

	millennialsCommaSeparatedList := os.Getenv("TEST_MILLENNIALS_EMAIL")
	fmt.Printf("Millennials comma separated list: %v\n", millennialsCommaSeparatedList)
	millennials := strings.Split(millennialsCommaSeparatedList, ", ")
	fmt.Printf("Millennials: %v\n", millennials)

	participants := append(babyBoomers, millennials...)

Start:
	fmt.Println("I'll be your Wichtel!")
	fmt.Println("Adding millennial's twice into the hat.")
	groupAMillennials := createSlipsOfPaper(millennials, "A")
	groupBMillennials := createSlipsOfPaper(millennials, "B")
	theHat := append(groupAMillennials, groupBMillennials...)

	// Shuffle the participants in the hat
	shuffleTheHat(theHat)

	wichtelMatches := make(map[string][]slipOfPaper, len(participants))
	for boomerIndex, boomer := range babyBoomers {
		var firstMatch, secondMatch slipOfPaper
		firstSlipOfPaperIndex := 0
		firstMatch = theHat[firstSlipOfPaperIndex]
		theHat = removeSlipOfPaperWithIndex(theHat, firstSlipOfPaperIndex)
		fmt.Printf("Found first match for baby boomer no. '%v'.\n", boomerIndex+1)
		for index, slipOfPaper := range theHat { // Iterate over remaining slips of paper in the hat
			secondMatch = slipOfPaper
			if secondMatch.email != firstMatch.email { // If the first and second match are not the same we have the result we want
				theHat = removeSlipOfPaperWithIndex(theHat, index) // overwrite hat removing the first entry
				fmt.Printf("Found second match for baby boomer no. '%v', that is not equal to the first match.\n", boomerIndex+1)
				break // we can break our of the for loop
			} // else we go to the next slip of paper
		}
		wichtelMatches[boomer] = []slipOfPaper{firstMatch, secondMatch}
	}

	fmt.Println("Adding baby boomers twice into the hat.")
	groupABabyBoomers := createSlipsOfPaper(babyBoomers, "A")
	groupBBabyBoomers := createSlipsOfPaper(babyBoomers, "B")
	theHat = append(theHat, groupABabyBoomers...)
	theHat = append(theHat, groupBBabyBoomers...)

	shuffleTheHat(theHat)

	// Find matches for the millennials
	for millennialIndex, millennial := range millennials {
		var firstMatch, secondMatch slipOfPaper
		// Make sure the participant did not pull their own name
		for index, slipOfPaper := range theHat {
			firstMatch = slipOfPaper
			if firstMatch.email != millennial {
				theHat = removeSlipOfPaperWithIndex(theHat, index)
				fmt.Printf("Found first match for millennial no. '%v'.\n", millennialIndex+1)
				break
			}
		}

		for index, slipOfPaper := range theHat { // Iterate over slips of paper in the hat starting at the second entry
			secondMatch = slipOfPaper
			// Make sure the participant did not pull their own name
			if secondMatch.email != millennial && secondMatch.email != firstMatch.email {
				theHat = removeSlipOfPaperWithIndex(theHat, index) // overwrite hat removing the first entry
				fmt.Printf("Found second match for millennial no. '%v', that is not equal to the first match.\n", millennialIndex+1)
				break // we can break our of the for loop
			}
		}
		wichtelMatches[millennial] = []slipOfPaper{firstMatch, secondMatch}
	}

	if len(theHat) > 0 {
		numberOfTries++
		if numberOfTries < 100 {
			fmt.Println("Oh-Oh, the last participant found their own name in the hat, starting over...\n")
			goto Start
		} else {
			log.Fatal("Tried 100 times to find a match for the last participant, aborting!")
		}
	}

	gmailService := setupGmailService()

	sendOutEmails(wichtelMatches, gmailService)

	deleteWichtelEmails(gmailService)
}

func createSlipsOfPaper(emails []string, group string) []slipOfPaper {
	var slipsOfPaper []slipOfPaper
	for _, email := range emails {
		slipsOfPaper = append(slipsOfPaper, slipOfPaper{
			email: email,
			group: group,
		})
	}
	return slipsOfPaper
}

func removeSlipOfPaperWithIndex(hat []slipOfPaper, index int) []slipOfPaper {
	newHat := make([]slipOfPaper, 0)
	newHat = append(newHat, hat[:index]...)   // Retrieve entries before index
	newHat = append(newHat, hat[index+1:]...) // Retrieve entries after index
	return newHat
}

func shuffleTheHat(hat []slipOfPaper) {
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

func sendOutEmails(wichtelMatches map[string][]slipOfPaper, gmailService *gmail.Service) {
	emailCounter := 0
	for wichtel, gifted := range wichtelMatches {
		err := sendEmail(
			wichtel,
			fmt.Sprintf("Congratulations, you are\nsecret santa '%s' to %s\nand secret santa '%s' to %s",
				gifted[0].group, gifted[0].email, gifted[1].group, gifted[1].email),
			gmailService)
		if err != nil {
			log.Fatal("Could not send email: ", err)
		}
		emailCounter++
		fmt.Printf("Sent out email no. %v\n", emailCounter)
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
		numberOfTries := 0
	TryToDelete:
		err = gmailService.Users.Messages.Delete("me", message.Id).Do()
		if err != nil {
			fmt.Printf("Could not delete message with ID '%v'; err: %v\n", message.Id, err)
			if numberOfTries < 10 {
				numberOfTries++
				time.Sleep(time.Second * time.Duration(numberOfTries))
				fmt.Printf("Trying again to delete message with ID '%v'\n", message.Id)
				goto TryToDelete
			} else {
				log.Fatal("Tried 10 times to delete and could not, aborting!")
			}

		}
		fmt.Printf("Deleted messages number %v\n", index+1)
	}
}
