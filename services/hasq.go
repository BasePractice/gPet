package services

import (
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"strings"
)

type Chain interface {
	Owned(key Key) int
	Key(passphrase string) Key
}

type Token interface {
	String() string
}

type Key interface {
	String() string
}

type Generator interface {
	String() string
}

type Owner interface {
	String() string
}

type token struct {
	data string
}

type key struct {
	data string
}

type generator struct {
	data string
}

type owner struct {
	data string
}

type chain struct {
	token    Token
	elements []element
}

func (c *chain) Key(passphrase string) Key {
	return &key{data: hash(len(c.elements), c.token.String(), passphrase)}
}

func (c *chain) Owned(key Key) int {
	next := len(c.elements)
	prev := c.elements[next-1]
	gen := CreateGenerator(next, c.token, key)
	prev.gen = gen
	if next >= 2 {
		third := c.elements[next-2]
		third.owner = CreateOwner(gen)
	}
	c.elements = append(c.elements, element{
		gen:   gen,
		key:   key,
		owner: createKeyOwner(key),
	})
	return next
}

type element struct {
	key   Key
	gen   Generator
	owner Owner
}

func (t token) String() string {
	return t.data
}

func (k key) String() string {
	return k.data
}

func (g generator) String() string {
	return g.data
}

func (o owner) String() string {
	return o.data
}

func CreateToken(data []byte) Token {
	d := sha3.Sum256(data)
	return &token{data: encodeToString(d[:])}
}

func createKey(token Token, passphrase string) Key {
	return &key{data: hash(0, token.String(), passphrase)}
}

func CreateGenerator(index int, token Token, key Key) Generator {
	return &generator{data: hash(index, token.String(), key.String())}
}

func CreateOwner(generator Generator) Owner {
	return &owner{data: hash(generator.String())}
}

func createKeyOwner(key Key) Owner {
	return &owner{data: key.String()}
}

func createKeyGenerator(key Key) Generator {
	return &generator{data: key.String()}
}

func CreateChain(token Token, passphrase string) Chain {
	key := createKey(token, passphrase)
	elements := make([]element, 0)
	elements = append(elements, element{
		key: key, owner: createKeyOwner(key),
		gen: createKeyGenerator(key),
	})
	return &chain{token: token, elements: elements}
}

func digest(params ...any) []byte {
	var h string
	for _, p := range params {
		h += fmt.Sprint(p)
	}
	d := sha3.Sum256([]byte(h))
	return d[:]
}

func hash(params ...any) string {
	return encodeToString(digest(params...))
}

func encodeToString(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

func decodeFromString(data string) []byte {
	ret, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	return ret
}
