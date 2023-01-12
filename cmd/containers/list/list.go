package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/configwizard/greenfinch-sdk/pkg/config"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"io/ioutil"
	"log"
	"os"
)

var (
	walletPath = flag.String("wallet", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
	containerID = flag.String("container", "", "specify the container")
)

func main() {

	ctx := context.Background()

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "error with flags")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *createWallet {
		secureWallet, err := wallet.GenerateNewSecureWallet(*walletPath, "some account label", *password)
		if err != nil {
			log.Fatal("error generating wallets", err)
		}
		file, _ := json.MarshalIndent(secureWallet, "", " ")
		_ = ioutil.WriteFile(*walletPath, file, 0644)
		log.Printf("created new wallets\r\n%+v\r\n", file)
		os.Exit(0)
	}
	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	userID := user.ID{}
	user.IDFromKey(&userID, key.PublicKey)

	// UserContainers implements neofs.NeoFS interface method.
	var prm pool.PrmContainerList
	prm.SetOwnerID(userID)

	fmt.Println("about to retrieve pool")
	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error retrieving pool %w", err)
	}

	r, err := pl.ListContainers(ctx, prm)
	if err != nil {
		fmt.Errorf("list user containers via connection pool: %w", err)
	}

	fmt.Println("user container list ", r)
}
