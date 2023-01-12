package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/tokens"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
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

	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	cnrID := cid.ID{}
	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("couldn't decode containerID")
	}

	//this doesn't feel correct??
	pKey := &keys.PrivateKey{PrivateKey: key}
	//todo: how do you attach a new session to a session Container?
	sc, err := tokens.BuildContainerSessionToken(pKey, 500, 500, 500, cid.ID{}, session.VerbContainerPut, *pKey.PublicKey())
	if err != nil {
		log.Fatal("error creating session token to create a container")
	}

	//for the time being, this is the same key
	specifiedTargetRole := eacl.NewTarget()
	eacl.SetTargetECDSAKeys(specifiedTargetRole, &key.PublicKey)

	var prm pool.PrmContainerSetEACL
	table, err := tokens.AllowKeyPutRead(cnrID, *specifiedTargetRole)
	if err != nil {
		log.Fatal("couldn't create eacl table", err)
	}

	prm.SetTable(table)
	//prm.SetWaitParams(x.await)

	if sc != nil {
		prm.WithinSession(*sc) //todo = what if the sc is nil? Why continue?
	}

	pl, err := gspool.GetPool(ctx, key)
	if err != nil {
		fmt.Errorf("%w", err)
	}

	if err := pl.SetEACL(ctx, prm); err != nil {
		fmt.Errorf("save eACL via connection pool: %w", err)
		return
	}
}

