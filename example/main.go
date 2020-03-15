package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
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

	accounts, _, err := client.Account.List()
	if err != nil {
		log.Fatal(err)
	}
	for i := range accounts {
		fmt.Printf("acc[%d]: %v\n\n", i, accounts[i])
		acc, _, err := client.Account.Get(accounts[i].ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Detailed account: %+v", acc)

		balance, _, err := client.Account.GetBalance(accounts[i].ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Balance: %+v", balance)

		statement, _, err := client.Account.GetStatement(accounts[i].ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Statement: %+v", statement)

		fees, _, err := client.Account.GetFees(accounts[i].ID, "")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("AllFees: %+v", fees)

		fee, _, err := client.Account.GetFees(accounts[i].ID, "internal_transfer")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Fee: %+v", fee)

		//Internal DryRun Transfer
		transfInput := types.TransferInput{
			AccountID: accounts[i].ID,
			Amount:    100,
			Target: types.Target{
				Account: types.TransferAccount{
					AccountCode: "334201",
				},
			},
		}

		idempotencyKey := uuid.New().String()
		transfer, _, err := client.Transfer.DryRunTransfer(transfInput, idempotencyKey)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Transfer(dry-run): %+v", transfer)

		//Internal Transfer
		transfer, _, err = client.Transfer.Transfer(transfInput, idempotencyKey)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Transfer: %+v", transfer)

		//External DryRun Transfer
		transfExtInput := types.TransferInput{
			AccountID: accounts[i].ID,
			Amount:    100,
			Target: types.Target{
				Account: types.TransferAccount{
					AccountCode:     "1234",
					BranchCode:      "7032",
					InstitutionCode: "001",
				},
				Entity: types.Entity{
					Name:         "James Bond",
					Document:     "00700700700",
					DocumentType: "cpf",
				},
			},
		}
		transfer, _, err = client.Transfer.DryRunTransfer(transfExtInput, idempotencyKey)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("External Transfer(dry-run): %+v", transfer)

		//External  Transfer
		transfer, _, err = client.Transfer.Transfer(transfExtInput, idempotencyKey)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("External Transfer: %+v", transfer)
	}
}
