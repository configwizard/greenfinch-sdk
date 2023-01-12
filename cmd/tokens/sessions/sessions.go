package sessions

import (
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// lifetimeOptions holds NeoFS epochs, iat -- epoch which the token was issued at, exp -- epoch when the token expires.
type lifetimeOptions struct {
	Iat uint64
	Exp uint64
}


func BuildSessionToken(key *keys.PrivateKey, lifetime lifetimeOptions, ctx sessionTokenContext, gateKey *keys.PublicKey) (*session.Container, error) {
	tok := new(session.Container)
	tok.ForVerb(ctx.verb)
	tok.AppliedTo(ctx.containerID)

	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(gateKey))

	tok.SetIat(lifetime.Iat)
	tok.SetNbf(lifetime.Iat)
	tok.SetExp(lifetime.Exp)

	return tok, tok.Sign(key.PrivateKey)
}
