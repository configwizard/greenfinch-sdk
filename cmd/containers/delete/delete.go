package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"io/ioutil"
	"log"
	"os"
)

var (
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
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

	cnrID := cid.ID{}

	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("couldn't decode containerID")
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

	//todo: how do you attach a new session to a session Container?
	sc := new(session.Container) //new session container for container actions

	var prm pool.PrmContainerDelete
	prm.SetContainerID(cnrID)

	if sc != nil {
		prm.SetSessionToken(*sc)
	}

	pl, err := gspool.GetPool(ctx, key)
	if err != nil {
		fmt.Errorf("%w", err)
	}

	if err := pl.DeleteContainer(ctx, prm); err != nil {
		fmt.Errorf("delete container via connection pool: %w", err)
	}
}
