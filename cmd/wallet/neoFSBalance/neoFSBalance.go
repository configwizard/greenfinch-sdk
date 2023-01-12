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

const usage = `Example

$ ./retrieveNeoFSBalance -wallets ./sample_wallets/wallet.rawContent.go
password is password
`

var (
	walletPath = flag.String("wallet", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")
)

func main() {
	fmt.Println(os.Getwd())
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

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


	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error %w", err)
	}
	userID := user.ID{}
	user.IDFromKey(&userID, key.PublicKey)
	blGet := pool.PrmBalanceGet{}
	blGet.SetAccount(userID)

	fmt.Println("waiting to retrieve result")
	res, err := pl.Balance(context.Background(), blGet)
	if err != nil {
		fmt.Errorf("error %w", err)
	}
	//res.Value()
	//var balance accounting.Decimal
	//
	//res.WriteToV2(&balance)
	fmt.Printf("Balance for %s: %v\n", userID, res.Value())
	//var prmInit client.PrmInit
	//prmInit.SetDefaultPrivateKey(key) // private key for request signing
	//prmInit.ResolveNeoFSFailures() // enable erroneous status parsing
	//
	//var c client.Client
	//c.Init(prmInit)
	//
	//var prmDial client.PrmDial
	//prmDial.SetServerURI("https://rpc1.morph.t5.fs.neo.org:51331") // endpoint address
	//
	//err = c.Dial(prmDial)
	//if err != nil {
	//	return
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	//defer cancel()
	//userID := user.ID{}
	//user.IDFromKey(&userID, key.PublicKey)
	//var prm client.PrmBalanceGet
	//prm.SetAccount(userID)
	//
	//res, err := c.BalanceGet(ctx, prm)
	//if err != nil {
	//	return
	//}
	//
	//fmt.Printf("Balance for %s: %v\n", userID, res.Amount())

}


