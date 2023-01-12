package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/configwizard/greenfinch-sdk/pkg/config"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/tokens"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	walletPath = flag.String("wallet", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")

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

	var containerAttributes = make(map[string]string)
	// First obtain client credentials: private key of request owner
	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}
	w := wallet.GetWalletFromPrivateKey(key)
	log.Println("using address ", w.Address)

	userID := user.ID{}
	user.IDFromKey(&userID, key.PublicKey)

	//this doesn't feel correct??
	pKey := &keys.PrivateKey{PrivateKey: key}
	//todo: how do you attach a new session to a session Container?

	placementPolicy := `REP 2 IN X 
	CBF 2
	SELECT 2 FROM * AS X
	`

	policy := netmap.PlacementPolicy{}
	if err := policy.DecodeString(placementPolicy); err != nil {
		fmt.Errorf("failed to build placement policy: %w", err)
		return
	}

	var cnr container.Container
	cnr.Init()
	cnr.SetPlacementPolicy(policy)
	cnr.SetOwner(userID)
	cnr.SetBasicACL(acl.PublicRWExtended)
	container.SetCreationTime(&cnr, time.Now())

	// todo: what is the difference between domain name and container name??
	var d container.Domain
	d.SetName("domain-name")

	container.WriteDomain(&cnr, d)
	container.SetName(&cnr, "container-name")

	for k, v := range containerAttributes {
		cnr.SetAttribute(k, v)
	}

	fmt.Println("about to retrieve pool")
	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error retrieving pool %w", err)
	}

	fmt.Println("pool retrieved")
	if err := pool.SyncContainerWithNetwork(ctx, &cnr, pl); err != nil {
		fmt.Errorf("sync container with the network state: %w", err)
		return
	}

	var prmPut pool.PrmContainerPut
	prmPut.SetContainer(cnr)

	iAt, exp, err := gspool.TokenExpiryValue(ctx, pl, 100)
	sc, err := tokens.BuildContainerSessionToken(pKey, iAt, iAt, exp, cid.ID{}, session.VerbContainerPut, *pKey.PublicKey())
	if err != nil {
		log.Fatal("error creating session token to create a container")
	}
	if sc != nil {
		prmPut.WithinSession(*sc)
	} else {
		//todo: what about just providing a key or a bearer token?
	}

	fmt.Println("about to put container")
	// send request to save the container
	idCnr, err := pl.PutContainer(ctx, prmPut) //see SetWaitParams to change wait times
	if err != nil {
		fmt.Println("save container via connection pool: %w", err)
		return
	}
	fmt.Println("container putted")
	fmt.Println("container created ", idCnr)
}
