package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const subjectID = "jul-wichtel-algorithm"

type slipOfPaper struct {
	email string
	group string
}

func main() {
	numberOfTries := 0

	babyBoomersCommaSeparatedList := os.Getenv("BABY_BOOMERS_EMAIL")
	babyBoomers := strings.Split(babyBoomersCommaSeparatedList, ", ")

	millennialsCommaSeparatedList := os.Getenv("MILLENNIALS_EMAIL")
	millennials := strings.Split(millennialsCommaSeparatedList, ", ")

	participants := append(babyBoomers, millennials...)

Start:
	// Create a fresh random generator for each attempt
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("I'll be your Wichtel!")
	fmt.Println("Adding millennial's twice into the hat.")
	groupAMillennials := createSlipsOfPaper(millennials, "A")
	groupBMillennials := createSlipsOfPaper(millennials, "B")
	theHat := append(groupAMillennials, groupBMillennials...)

	// Shuffle the participants in the hat
	shuffleTheHat(theHat, rng)

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

	shuffleTheHat(theHat, rng)

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

	// Post-processing: Ensure special millennial gets special boomer (if env vars are set)
	specialMillenial := os.Getenv("SPECIAL_MILLENNIAL_EMAIL")
	specialBoomer := os.Getenv("SPECIAL_BOOMER_EMAIL")

	if specialMillenial != "" && specialBoomer != "" {
		fmt.Printf("Special pairing environment variables detected - applying custom pairing logic\n")

		specialBoomerExists := contains(babyBoomers, specialBoomer)
		specialMillenialExists := contains(millennials, specialMillenial)

		if specialBoomerExists && specialMillenialExists {
			fmt.Printf("Checking for special pairing between designated participants\n")

			// Check if special millennial already has the special boomer
			specialMillenialMatches := wichtelMatches[specialMillenial]
			var specialMillenialHasBoomer bool
			var specialMillenialBoomerIndex int

			for i, match := range specialMillenialMatches {
				if match.email == specialBoomer {
					specialMillenialHasBoomer = true
					specialMillenialBoomerIndex = i
					break
				}
			}

			if specialMillenialHasBoomer {
				fmt.Printf("✅ Special millennial already has special boomer at index %d - no swap needed!\n", specialMillenialBoomerIndex)
			} else {
				// Find who currently has the special boomer
				var currentOwnerOfSpecialBoomer string
				var specialBoomerIndex int
				var foundSpecialBoomer bool

				for participant, matches := range wichtelMatches {
					for i, match := range matches {
						if match.email == specialBoomer {
							currentOwnerOfSpecialBoomer = participant
							specialBoomerIndex = i
							foundSpecialBoomer = true
							break
						}
					}
					if foundSpecialBoomer {
						break
					}
				}

				if foundSpecialBoomer {
					fmt.Printf("Special boomer is assigned to different participant, performing swap\n")

					// Get the matches for both participants
					currentOwnerMatches := wichtelMatches[currentOwnerOfSpecialBoomer]

					// Find what the current owner has left (the other match, not the special boomer)
					otherIndexForCurrentOwner := 1 - specialBoomerIndex
					currentOwnerOtherMatch := currentOwnerMatches[otherIndexForCurrentOwner]

					// Find which of the special millennial's matches we should give to avoid duplicates
					var swapIndex int
					if specialMillenialMatches[0].email != currentOwnerOtherMatch.email {
						swapIndex = 0
					} else if specialMillenialMatches[1].email != currentOwnerOtherMatch.email {
						swapIndex = 1
					} else {
						fmt.Printf("⚠️  Cannot perform swap - would create duplicate for current owner\n")
						return
					}

					// Perform the swap: give special millennial the special boomer,
					// and give current owner the non-duplicate match from special millennial
					temp := specialMillenialMatches[swapIndex]
					wichtelMatches[specialMillenial][swapIndex] = currentOwnerMatches[specialBoomerIndex]
					wichtelMatches[currentOwnerOfSpecialBoomer][specialBoomerIndex] = temp

					fmt.Printf("✅ Successfully swapped! Special millennial now gets special boomer as %s\n", wichtelMatches[specialMillenial][swapIndex].group)
				} else {
					fmt.Printf("⚠️  Could not find special boomer in the matches\n")
				}
			}
		} else {
			if !specialBoomerExists {
				fmt.Printf("⚠️  Special boomer not found in baby boomers list\n")
			}
			if !specialMillenialExists {
				fmt.Printf("⚠️  Special millennial not found in millennials list\n")
			}
		}
	} else {
		fmt.Println("No special pairing environment variables set - running normal algorithm")
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

func shuffleTheHat(hat []slipOfPaper, rng *rand.Rand) {
	rng.Shuffle(len(hat), func(i, j int) {
		hat[i], hat[j] = hat[j], hat[i]
	})
}

// Helper function to check if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
	config, err := google.ConfigFromJSON([]byte(credentialsJSON), "https://mail.google.com/")
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
