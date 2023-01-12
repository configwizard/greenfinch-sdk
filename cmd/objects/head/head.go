package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/configwizard/greenfinch-sdk/pkg/config"
	gspool "github.com/configwizard/greenfinch-sdk/pkg/pool"
	"github.com/configwizard/greenfinch-sdk/pkg/tokens"
	"github.com/configwizard/greenfinch-sdk/pkg/wallet"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"io/ioutil"
	"log"
	"os"

	//"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
)

func isErrAccessDenied(err error) (string, bool) {
	unwrappedErr := errors.Unwrap(err)
	for unwrappedErr != nil {
		err = unwrappedErr
		unwrappedErr = errors.Unwrap(err)
	}
	switch err := err.(type) {
	default:
		return "", false
	case apistatus.ObjectAccessDenied:
		return err.Reason(), true
	case *apistatus.ObjectAccessDenied:
		return err.Reason(), true
	}
}

// todo: BINGO: https://github.com/nspcc-dev/neofs-s3-gw/blob/master/internal/neofs/neofs.go

//// DownloadByAddress handles download requests using simple cid/oid format.
//func (d *Downloader) DownloadByAddress(c *fasthttp.RequestCtx) {
//	d.byAddress(c, request.receiveFile)
//}

// byAddress is a wrapper for function (e.g. request.headObject, request.receiveFile) that
// prepares request and object address to it.
func main() {

	ctx := context.Background()
	var (
		walletPath   = flag.String("wallet", "", "path to JSON wallets file")
		walletAddr   = flag.String("address", "", "wallets address [optional]")
		createWallet = flag.Bool("create", false, "create a wallets")
		password     = flag.String("password", "", "wallet password")
		containerID  = flag.String("container", "", "specify the container")
		objectID = flag.String("object", "", "specify the object")
	)

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
	cnrID := cid.ID{}

	if err := cnrID.DecodeString(*containerID); err != nil {
		log.Fatal("couldn't decode containerID")
	}

	objID := oid.ID{}
	if err = objID.DecodeString(*objectID); err != nil {
		fmt.Println("wrong object id", err)
		return
	}

	var addr oid.Address
	addr.SetContainer(cnrID)
	addr.SetObject(objID)

	var prmHead pool.PrmObjectHead
	prmHead.SetAddress(addr)

	//this doesn't feel correct??
	pKey := &keys.PrivateKey{PrivateKey: key}

	config := config.ReadConfig()
	pl, err := gspool.GetPool(ctx, key, config.Peers)
	if err != nil {
		fmt.Errorf("error retrieving pool %w", err)
	}


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
	if bt != nil {
		fmt.Println("using bearer token")
		prmHead.UseBearer(*bt)
	} else {
		prmHead.UseKey(&key)
	}
	hdr, err := pl.HeadObject(ctx, prmHead)
	if err != nil {
		if reason, ok := isErrAccessDenied(err); ok {
			fmt.Printf("%w: %s\r\n", err, reason)
			return
		}
		fmt.Errorf("read object header via connection pool: %w", err)
		return
	}

	for _, attr := range hdr.Attributes() {
		key := attr.Key()
		val := attr.Value()
		fmt.Println(key, val)
		switch key {
		case object.AttributeFileName:
		case object.AttributeTimestamp:
		case object.AttributeContentType:
		}
	}

	fmt.Printf("%+v\r\n", hdr.PayloadSize())


}
