# go-stone-openbank

A Go library to connect with [Stone Open Banking API](https://docs.openbank.stone.com.br/)

## How to install

```sh
  $ go get github.com/stone-co/go-stone-openbank
```

## Example Usage

```go
package main

import (
	openbank "github.com/stone-co/go-stone-openbank"
	"github.com/stone-co/go-stone-openbank/types"
)

func main() {
	clientID := os.Getenv("STONE_CLIENT_ID")
	privKeyPath := os.Getenv("STONE_PRIVATE_KEY")

	client := openbank.NewClient(
		openbank.WithClientID(clientID),
		openbank.SetPrivateKey(privKeyPath),
		openbank.UseSandbox(),
	//	openbank.EnableDebug(),
	)

	err := client.Authenticate()
	if err != nil {
		log.Fatal(err)
	}

	consentLink, err := client.ConsentLink()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("consent_link: %s\n", consentLink)

	accounts, _, err := client.Account.List()
	if err != nil {
		log.Fatal(err)
	}
	for i := range accounts {
		balance, _, err := client.Account.GetBalance(accounts[i].ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Balance: %+v", balance)
  }
}
```

see full [example](https://github.com/stone-co/go-stone-openbank/blob/master/example/main.go)
