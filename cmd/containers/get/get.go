package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/configwizard/greenfinch-sdk/pkg/config"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
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

//integration tests here for reference https://github.com/nspcc-dev/neofs-http-gw/blob/278376643a46db50d6e04cc73d853f30fc1f3708/integration_test.go

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
	w := wallet.GetWalletFromPrivateKey(key)
	log.Println("using address ", w.Address)

	userID := user.ID{}
	user.IDFromKey(&userID, key.PublicKey)

	fmt.Println("about to retrieve pool")
	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error retrieving pool %w", err)
	}

	cnrID := cid.ID{}
	cnrID.DecodeString(*containerID)
	var prmGet pool.PrmContainerGet
	prmGet.SetContainerID(cnrID)

	fmt.Println("about to get container")
	// send request to save the container
	cnr, err := pl.GetContainer(ctx, prmGet) //see SetWaitParams to change wait times
	if err != nil {
		fmt.Println("save container via connection pool: %w", err)
		return
	}
	byteData, err := cnr.MarshalJSON()
	fmt.Print("container retreived %+v -- %s\r\n", string(byteData), err)
}
