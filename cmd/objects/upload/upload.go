package main

import (
	"bytes"
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
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
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
	fn := "upload.gif"

	//expiration epoch https://github.com/nspcc-dev/neofs-s3-gw/blob/master/internal/neofs/neofs.go#L542
	var filtered = make(map[string]string)
	attributes := make([]object.Attribute, 0, len(filtered))
	// prepares attributes from filtered headers
	for key, val := range filtered {
		attribute := object.NewAttribute()
		attribute.SetKey(key)
		attribute.SetValue(val)
		attributes = append(attributes, *attribute)
	}
	if _, ok := filtered[object.AttributeFileName]; !ok {
		filename := object.NewAttribute()
		filename.SetKey(object.AttributeFileName)
		filename.SetValue(fn)
		attributes = append(attributes, *filename)
	}
	if _, ok := filtered[object.AttributeTimestamp]; !ok {
		timestamp := object.NewAttribute()
		timestamp.SetKey(object.AttributeTimestamp)
		timestamp.SetValue(strconv.FormatInt(time.Now().Unix(), 10))
		attributes = append(attributes, *timestamp)
	}
	//this doesn't feel correct??
	pKey := &keys.PrivateKey{PrivateKey: key}

	//bt = nil
	//todo set the bearer token properties to upload this object
	obj := object.New()
	obj.SetContainerID(cnrID)
	obj.SetOwnerID(&userID)
	obj.SetAttributes(attributes...)
	data := []byte("this is some data stored as a byte slice in Go Lang!")
	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	obj.SetPayloadSize(uint64(len(data)))


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
	var prm pool.PrmObjectPut
	prm.SetHeader(*obj)
	prm.SetPayload(reader)
	if bt != nil {
		fmt.Println("using bearer token")
		prm.UseBearer(*bt)
	} else {
		prm.UseKey(&key)
	}
	if idObj, err := pl.PutObject(ctx, prm); err != nil {
		reason, ok := isErrAccessDenied(err)
		if ok {
			fmt.Printf("%w: %s\r\n", err, reason)
			return
		}
		fmt.Println("save object via connection pool: %s", err)
		return
	} else {
		fmt.Println("created object ", idObj, " in container ", cnrID)
	}
}
