package main

import (
	"io/ioutil"
	"log"
	"os"

	openbank "github.com/stone-co/go-stone-openbank"
)

func main() {
	clientID := os.Getenv("STONE_CLIENT_ID")
	privKeyPath := os.Getenv("STONE_PRIVATE_KEY")
	consentURL := os.Getenv("STONE_CONSENT_REDIRECT_URL")
	baseUrl := os.Getenv("API_BASE_URL")

	pemPrivKey := readFileContent(privKeyPath)
	baseUrlOpt, err := openbank.SetBaseURL(baseUrl)
	if err != nil {
		log.Fatal(err)
	}

	client, err := openbank.NewClient(
		openbank.WithClientID(clientID),
		openbank.WithPEMPrivateKey(pemPrivKey),
		openbank.SetConsentURL(consentURL),
		openbank.UseSandbox(),
		baseUrlOpt,
		//	openbank.EnableDebug(),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Authenticate(); err != nil {
		log.Fatal(err)
	}

	request, err := client.NewAPIRequest("GET", "/example", nil)
	if err != nil {
		log.Fatal(err)
	}

	successResponse := struct {
		Message string `json:"message,omitempty"`
	}{}
	errorResponse := struct {
		Message string `json:"message,omitempty"`
		Detail  string `json:"detail,omitempty"`
	}{}

	response, err := client.Do(request, successResponse, errorResponse)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println(response)
}

func readFileContent(path string) []byte {
	content, _ := ioutil.ReadFile(path)
	return content
}
