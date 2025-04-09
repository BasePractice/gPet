package services

import "testing"

func TestHashQ_CreateKey(t *testing.T) {
	tok := CreateToken([]byte("DATA"))
	key := createKey(tok, "password")
	t.Log(key)
}

func TestHashQ_CreateChain(t *testing.T) {
	chain := CreateChain([]byte("DATA"), "password")
	appendOwner(chain, "password1")
	appendOwner(chain, "password2")
	appendOwner(chain, "password3")
	appendOwner(chain, "password4")
	appendOwner(chain, "password5")
	appendOwner(chain, "password6")
	t.Log(chain)
	t.Log(chain.Validate())
}

func appendOwner(ch Chain, passphrase string) {
	key := ch.Key(passphrase)
	ch.Owned(key)
}
