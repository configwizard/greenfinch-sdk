package tokens

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// see here if you want to convert a time to an epoch https://github.com/nspcc-dev/neofs-s3-gw/blob/master/internal/neofs/neofs.go


func BuildBearerToken(key *keys.PrivateKey, table *eacl.Table, lIat, lNbf, lExp uint64, gateKey *keys.PublicKey) (*bearer.Token, error) {
	var userID user.ID
	user.IDFromKey(&userID, (ecdsa.PublicKey)(*gateKey)) //my understanding is the gateKey is who you want to be able to use this key to access containers?

	var bearerToken bearer.Token
	//i understand this will restrict everything to the 'other' accounts
	for _, r := range restrictedRecordsForOthers() {
		table.AddRecord(r)
	}

	bearerToken.SetEACLTable(*table)
	bearerToken.ForUser(userID)
	bearerToken.SetExp(lExp)
	bearerToken.SetIat(lIat)
	bearerToken.SetNbf(lNbf)

	err := bearerToken.Sign(key.PrivateKey) //is this the owner who is giving access priveliges???
	if err != nil {
		return nil, fmt.Errorf("sign bearer token: %w", err)
	}

	return &bearerToken, nil
}