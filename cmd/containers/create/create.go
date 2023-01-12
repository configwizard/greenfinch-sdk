package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
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
	walletPath = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password = flag.String("password", "", "wallet password")

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

	//todo: how do you attach a new session to a session Container?
	sc := new(session.Container) //new session container for container actions

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
	d.SetName("[domain] todo - extract this as a variable")

	container.WriteDomain(&cnr, d)
	container.SetName(&cnr, "[container] todo - extract this as a variable")

	for k, v := range containerAttributes {
		cnr.SetAttribute(k, v)
	}

	pl, err := gspool.GetPool(ctx, key)
	if err != nil {
		fmt.Errorf("%w", err)
	}

	if err := pool.SyncContainerWithNetwork(ctx, &cnr, pl); err != nil {
		fmt.Errorf("sync container with the network state: %w", err)
		return
	}

	var prmPut pool.PrmContainerPut
	prmPut.SetContainer(cnr)

	if sc != nil {
		prmPut.WithinSession(*sc)
	} else {
		//todo: what about just providing a key or a bearer token?
	}

	// send request to save the container
	idCnr, err := pl.PutContainer(ctx, prmPut) //see SetWaitParams to change wait times
	if err != nil {
		fmt.Errorf("save container via connection pool: %w", err)
		return
	}
	fmt.Println("container created ", idCnr)
}
