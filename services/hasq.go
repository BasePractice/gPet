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
	Validate() bool
}

type Token interface {
	String() string
}

type Key interface {
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
	elements []*element
}

func (c *chain) Validate() bool {
	for i := len(c.elements) - 1; i >= 1; i-- {
		var curr = c.elements[i]
		var prev = c.elements[i-1]
		var gen = createGenerator(i, c.token, curr.key)
		if gen.data != prev.gen.data {
			return false
		}
		if i-2 >= 0 && curr.gen != nil {
			var own = createOwner(curr.gen)
			if own.data != prev.owner.data {
				return false
			}
		}
	}
	return true
}

func (c *chain) Key(passphrase string) Key {
	return &key{data: hash(len(c.elements), c.token.String(), passphrase)}
}

func (c *chain) Owned(key Key) int {
	next := len(c.elements)
	prev := c.elements[next-1]
	gen := createGenerator(next, c.token, key)
	prev.gen = gen
	if next >= 2 {
		third := c.elements[next-2]
		third.owner = createOwner(gen)
	}
	c.elements = append(c.elements, &element{
		gen:   nil,
		key:   key,
		owner: nil,
	})
	return next
}

type element struct {
	key   Key
	gen   *generator
	owner *owner
}

func (t token) String() string {
	return t.data
}

func (k key) String() string {
	return k.data
}

func CreateToken(data []byte) Token {
	d := sha3.Sum256(data)
	return &token{data: encodeToString(d[:])}
}

func createKey(token Token, passphrase string) Key {
	return &key{data: hash(0, token.String(), passphrase)}
}

func createGenerator(index int, token Token, key Key) *generator {
	return &generator{data: hash(index, token.String(), key.String())}
}

func createOwner(generator *generator) *owner {
	return &owner{data: hash(generator.data)}
}

func createKeyOwner(key Key) *owner {
	return &owner{data: key.String()}
}

func createKeyGenerator(key Key) *generator {
	return &generator{data: key.String()}
}

func CreateChain(tokenData []byte, passphrase string) Chain {
	tok := CreateToken(tokenData)
	k := createKey(tok, passphrase)
	elements := make([]*element, 0)
	elements = append(elements, &element{
		key: k, owner: createKeyOwner(k),
		gen: createKeyGenerator(k),
	})
	return &chain{token: tok, elements: elements}
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
