# merchant-go-stone-openbank

A forked Go library of [go-stone-openbank](https://github.com/stone-payments/go-stone-openbank) to connect [Stone Open Banking API](https://docs.openbank.stone.com.br/)

## How to install

```sh
go get github.com/stone-payments/merchant-go-stone-openbank/v3
```

## Example Usage

```go
package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	openbank "github.com/stone-payments/merchant-go-stone-openbank/v3"
)

func main() {
	clientID := os.Getenv("STONE_CLIENT_ID")
	privKeyPath := os.Getenv("STONE_PRIVATE_KEY")
	consentURL := os.Getenv("STONE_CONSENT_REDIRECT_URL")

	pemPrivKey := readFileContent(privKeyPath)

	client, err := openbank.NewClient(
		openbank.WithClientID(clientID),
		openbank.SetConsentURL(consentURL),
		openbank.WithPEMPrivateKey(pemPrivKey),
		openbank.UseSandbox(),
		//	openbank.EnableDebug(),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Authenticate(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func readFileContent(path string) []byte {
	content, _ := ioutil.ReadFile(path)
	return content
}
```

see full [example](https://github.com/stone-payments/merchant-go-stone-openbank/blob/master/example/main.go)
