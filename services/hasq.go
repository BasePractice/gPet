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
	Equal(t Token) bool
}

type Key interface {
	String() string
}

type Generator interface {
	String() string
	Equal(g Generator) bool
}

type Owner interface {
	String() string
	Equal(o Owner) bool
}

type token struct {
	data string
}

func (t token) Equal(other Token) bool {
	switch other.(type) {
	case token:
		return t.data == other.(token).data
	case *token:
		return t.data == other.(*token).data
	default:
		panic(fmt.Sprintf("invalid token type: %T", other))
	}
}

type key struct {
	data string
}

type generator struct {
	data string
}

func (g generator) Equal(other Generator) bool {
	switch other.(type) {
	case generator:
		return g.data == other.(generator).data
	case *generator:
		return g.data == other.(*generator).data
	default:
		panic(fmt.Sprintf("invalid generator type: %T", other))
	}
}

type owner struct {
	data string
}

func (o owner) Equal(other Owner) bool {
	switch other.(type) {
	case owner:
		return o.data == other.(owner).data
	case *owner:
		return o.data == other.(*owner).data
	default:
		panic(fmt.Sprintf("invalid owner type: %T", other))
	}
}

type chain struct {
	token    Token
	elements []*element
}

func (c *chain) Validate() bool {
	for i := len(c.elements) - 1; i >= 1; i-- {
		var curr = c.elements[i]
		var prev = c.elements[i-1]
		var genn = createGenerator(i, c.token, curr.key)
		if !genn.Equal(prev.gen) {
			return false
		}
		if i-2 >= 0 && curr.gen != nil {
			var ownr = createOwner(curr.gen)
			if !ownr.Equal(prev.owner) {
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
	gen   Generator
	owner Owner
}

func (e element) String() string {
	return e.key.String() + ":" + e.gen.String() + ":" + e.owner.String()
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

func createGenerator(index int, token Token, key Key) Generator {
	return &generator{data: hash(index, token.String(), key.String())}
}

func createOwner(generator Generator) Owner {
	return &owner{data: hash(generator.String())}
}

func createKeyOwner(key Key) Owner {
	return &owner{data: key.String()}
}

func createKeyGenerator(key Key) Generator {
	return &generator{data: key.String()}
}

func CreateChain(tokenData []byte, passphrase string) Chain {
	token := CreateToken(tokenData)
	key := createKey(token, passphrase)
	elements := make([]*element, 0)
	elements = append(elements, &element{
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
