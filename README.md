# Greenfinch SDK


## Examples

In the `cmd` directory is a series of client examples that should be very straightforward to get working. The 
first/only thing required will be to generate a key. To do so, each example offers a flag that will genrate you a 
key if you need one. However don't forget to go to the testnet faucet to load up with (testnet) gas.

### Running examples.

You will ultimately need a wallet with testnet (T5) gas available. At that point you will need to have transferred 
gas to the NeoFS smart contract to pay for interactivity.

You can use any of the examples to create a wallet, the [testnet fauct here](https://n3t5wish.ngd.network/#/) to 
receive Neo and Gas and then the `cmd/wallet/makeTransaction/makeTransaction.go` example to make a transaction to 
the NeoFS contract.

After that, all examples can be run with 

```shell
go run XXX.go --wallet ../mywallets --password abc123
```
where XXX.go is the example you are trying to run.

### Issues 

Below are a list of issues/questions as to the examples/issues I am not sure of at this time.

#### 1. Pools

From looking at all the examples/code available I should not be using a client to make container/object 
(put/get/delete/search) but I should be using a pool. [Please see pool.go](pkg/pool/pool.go)

In here I believe that for the pool to work

1. I need to provide a list of nodes that the pool can connect to, where can I get these? Can I retrieve the 
   dynamically at runtime?
2. Pool needs a private key. In my use case, the users will be the key. However I want them to be able to use 
   walletConnect. Can I start a pool based on the response from WalletConnect

#### 2. SessionTokens

`pkg/tokens`

My understanding is SessionTokens are required during creating/deletion/listing of containers. BearerTokens cannot 
be used.

1. How can I create a sessionToken with the output of WalletConnect?
2. How does a SessionToken relate to a `session.Object` or a `session.Container` ? Are these just wrappers around a 
   token?

#### 3. BearerTokens

`pkg/tokens`

1. Both session tokens and bearer tokens can have 'gate keys'. My understanding was that session tokens could only 
   be issued for the 'owner' and not on behalf of someone else and gate keys therefore would only make sense for 
   bearer tokens?
2. Am I right in thinking that passing an unsigned bearer token to wallet Connect will allow walletConnect to sign 
   it on behalf of a user who is being given privileges to then use for the duration the bearer token is valid? Is 
   that correct?

#### 4. Setting the Iat, Nbg, Exp

I used to have a function that would generate these for me, however it uses a client from the old days... Is there a 
way I can generate realistic/practical values for these based on current epoch and a good "expiry" for tokens?
