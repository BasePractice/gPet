package services

import "testing"

func TestHashQ_CreateKey(t *testing.T) {
	tok := CreateToken([]byte("DATA"))
	key := createKey(tok, "password")
	t.Log(key)
}

func TestHashQ_CreateChain(t *testing.T) {
	tok := CreateToken([]byte("DATA"))
	chain := CreateChain(tok, "password")
	t.Log(chain)
	key := chain.Key("password2")
	t.Log(key)
	chain.Owned(key)
	key = chain.Key("password3")
	t.Log(key)
	chain.Owned(key)
	t.Log(chain)
}
