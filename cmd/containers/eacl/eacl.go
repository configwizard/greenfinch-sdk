package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
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

	key, err := wallet.GetCredentialsFromPath(*walletPath, *walletAddr, *password)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	cnrID := cid.ID{}
	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("couldn't decode containerID")
	}

	sc := new(session.Container) //new session container for container actions

	//for the time being, this is the same key
	specifiedTargetRole := eacl.NewTarget()
	eacl.SetTargetECDSAKeys(specifiedTargetRole, &key.PublicKey)

	var prm pool.PrmContainerSetEACL
	table, err := AllowKeyPutRead(cnrID, *specifiedTargetRole)
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

//AllowOthersReadOnly from https://github.com/nspcc-dev/neofs-s3-gw/blob/fdc07b8dc15272e2aabcbd7bb8c19e435c94e392/authmate/authmate.go#L358
func AllowKeyPutRead(cid cid.ID, toWhom eacl.Target) (eacl.Table, error) {
	table := eacl.Table{}
	targetOthers := eacl.NewTarget()
	targetOthers.SetRole(eacl.RoleOthers)

	getAllowRecord := eacl.NewRecord()
	getAllowRecord.SetOperation(eacl.OperationGet)
	getAllowRecord.SetAction(eacl.ActionAllow)
	getAllowRecord.SetTargets(toWhom)

	//getDenyRecord := eacl.NewRecord()
	//getDenyRecord.SetOperation(eacl.OperationGet)
	//getDenyRecord.SetAction(eacl.ActionDeny)
	//getDenyRecord.SetTargets(toWhom)

	putAllowRecord := eacl.NewRecord()
	putAllowRecord.SetOperation(eacl.OperationPut)
	putAllowRecord.SetAction(eacl.ActionAllow)
	putAllowRecord.SetTargets(toWhom)

	//putDenyRecord := eacl.NewRecord()
	//putDenyRecord.SetOperation(eacl.OperationPut)
	//putDenyRecord.SetAction(eacl.ActionDeny)
	//putDenyRecord.SetTargets(toWhom)

	table.SetCID(cid)
	table.AddRecord(getAllowRecord)
	table.AddRecord(putAllowRecord)
	//table.AddRecord(getDenyRecord)
	//table.AddRecord(putDenyRecord)
	//table.AddRecord(denyGETRecord)//deny must come first. Commented while debugging

	return table, nil
}
