package pool

import (
	"context"
	"crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"time"
)


/*
questions:
1. do i need to provide the URLs for connections in the pool
2. Whats the difference between a pool and a client and which should I use?
3. Can a pool be created without knowing the private key (wallet connect)?
	- if not, do I think use a client? I cant work out how to make the requests (put/get/delte) on a client
 */
func GetPool(ctx context.Context, key ecdsa.PrivateKey) (*pool.Pool, error) {
	var prm pool.InitParameters

	prm.AddNode() //how do I add nodes? Can this happen automatically?

	prm.SetNodeDialTimeout(1 * time.Minute)

	prm.SetNodeStreamTimeout(1 * time.Minute)

	prm.SetHealthcheckTimeout(1 * time.Minute)

	prm.SetClientRebalanceInterval(1 * time.Minute)

	prm.SetErrorThreshold(1)
	prm.SetKey(&key)
	//does this need setting or does this have a default?
	//prm.SetSessionExpirationDuration(10)
	p, err := pool.NewPool(prm)
	if err != nil {
		return p, err
	}

	if err = p.Dial(ctx); err != nil {
		return p, err
	}

	return p, nil
}
