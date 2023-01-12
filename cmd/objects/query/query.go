package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"io/ioutil"
	"log"
	"os"
)

var (
	walletPath   = flag.String("wallets", "", "path to JSON wallets file")
	walletAddr   = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password     = flag.String("password", "", "wallet password")
	containerID  = flag.String("container", "", "specify the container")
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

	pl, err := gspool.GetPool(ctx, key)
	if err != nil {
		fmt.Errorf("%w", err)
	}

	cnrID := cid.ID{}
	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("can't create container ID:", err)
	}

	var bt = new(bearer.Token)
	var sc = new(session.Object) //what is session.Object vs session.Contaner vs session.Token?
	//and what do we need to do to a session object to 'validate' the request
	prms := pool.PrmObjectSearch{}
	if bt != nil {
		prms.UseBearer(*bt)
	} else if sc != nil{
		prms.UseSession(*sc)
	} else {
		prms.UseKey(&key)
	}

	prms.SetContainerID(cnrID)

	filter := object.SearchFilters{}
	filter.AddRootFilter()

	prms.SetFilters(filter)
	objects, err := pl.SearchObjects(ctx, prms)
	if err != nil {
		return
	}
	var list []oid.ID
	err = objects.Iterate(func(id oid.ID) bool {
		list = append(list, id)
		return false
	})
	fmt.Printf("%+v %s\r\n", list, err)
}
