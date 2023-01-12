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
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"io/ioutil"
	"log"
	"os"
)

var (
	walletPath   = flag.String("wallet", "", "path to JSON wallets file")
	walletAddr   = flag.String("address", "", "wallets address [optional]")
	createWallet = flag.Bool("create", false, "create a wallets")
	password     = flag.String("password", "", "wallet password")
	containerID  = flag.String("container", "", "specify the container")
)

//see here for search by attribute vs search by address https://github.com/nspcc-dev/neofs-http-gw/blob/master/downloader/download.go#L340
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
	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error retrieving pool %w", err)
	}

	cnrID := cid.ID{}
	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("can't create container ID:", err)
	}

	//this doesn't feel correct??
	pKey := &keys.PrivateKey{PrivateKey: key}
	target := eacl.Target{}
	target.SetRole(eacl.RoleUser)
	target.SetBinaryKeys([][]byte{pKey.Bytes()})
	table, err := tokens.AllowKeyPutRead(cnrID, target)
	if err != nil {
		log.Fatal("error retrieving table ", err)
	}
	iAt, exp, err := gspool.TokenExpiryValue(ctx, pl, 100)
	bt, err := tokens.BuildBearerToken(pKey, &table, iAt, iAt, exp, pKey.PublicKey())
	if err != nil {
		log.Fatal("error creating bearer token to upload object")
	}

	prms := pool.PrmObjectSearch{}
	if bt != nil{
		prms.UseBearer(*bt)
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
	if err = objects.Iterate(func(id oid.ID) bool {
		list = append(list, id)
		return false
	}); err != nil {
		log.Fatalf("error listing objects %s\r\n", err)
	}
	fmt.Printf("list objects %+v\r\n", list)
}
