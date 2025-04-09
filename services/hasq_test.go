package services

import "testing"

func TestHashQ_CreateKey(t *testing.T) {
	tok := CreateToken([]byte("DATA"))
	k := createKey(tok, "password")
	t.Log(k)
}

func TestHashQ_CreateChain(t *testing.T) {
	ch := CreateChain([]byte("DATA"), "password")
	appendOwner(ch, "password1")
	appendOwner(ch, "password2")
	appendOwner(ch, "password3")
	appendOwner(ch, "password4")
	appendOwner(ch, "password5")
	appendOwner(ch, "password6")
	t.Log(ch)
	t.Log(ch.Validate())
}

func appendOwner(ch Chain, passphrase string) {
	k := ch.Key(passphrase)
	ch.Owned(k)
}
