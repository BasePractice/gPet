package services

import (
	"pet/hasqchain"
)

// Chain is an alias for hasqchain.Chain
type Chain = hasqchain.Chain

// Change is an alias for hasqchain.Change
type Change = hasqchain.Change

// Token is an alias for hasqchain.Token
type Token = hasqchain.Token

// Key is an alias for hasqchain.Key
type Key = hasqchain.Key

// LoadKey creates a key from a hash.
func LoadKey(hash string) Key {
	return hasqchain.LoadKey(hash)
}

// CreateToken creates a token from data.
func CreateToken(data []byte) Token {
	return hasqchain.CreateToken(data)
}

// CreateChain creates a new chain with the given token data and passphrase.
func CreateChain(tokenData []byte, passphrase string) Chain {
	return hasqchain.CreateChain(tokenData, passphrase)
}

// CreateEmptyChain creates an empty chain with the given token and length.
func CreateEmptyChain(tok string, length uint64) Chain {
	return hasqchain.CreateEmptyChain(tok, length)
}
