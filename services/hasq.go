package services

import (
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"strings"
)

type Chain interface {
	Owned(key Key) Change
	GetOwner() (uint64, Key)
	Key(passphrase string) (uint64, Key)
	KeyOn(id uint64, passphrase string) (uint64, Key)
	Validate() bool
	Push(id uint64, key string, gen *string, owner *string) error
}

type Change struct {
	N     uint64
	Gen   *string
	GenId uint64
	Own   *string
	OwnId uint64
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
	length   uint64
	token    Token
	elements []*element
}

func (c *chain) KeyOn(id uint64, passphrase string) (uint64, Key) {
	return id, &key{data: hash(id, c.token.String(), passphrase)}
}

func (c *chain) GetOwner() (uint64, Key) {
	l := len(c.elements)
	if l == 0 {
		return 0, nil
	}
	return c.elements[l-1].id, c.elements[l-1].key
}

func (c *chain) Push(id uint64, k string, gen *string, ow *string) error {
	var g *generator = nil
	var o *owner = nil

	if gen != nil {
		g = &generator{data: *gen}
	}
	if ow != nil {
		o = &owner{data: *ow}
	}

	c.elements = append(c.elements, &element{
		id:    id,
		gen:   g,
		key:   &key{data: k},
		owner: o,
	})
	return nil
}

func (c *chain) Validate() bool {
	if len(c.elements) == 0 {
		return true
	}
	for i := uint64(len(c.elements)) - 1; i >= 1; i-- {
		var curr = c.elements[i]
		var prev = c.elements[i-1]
		var gen = createGenerator(curr.id, c.token, curr.key)
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

func (c *chain) Key(passphrase string) (uint64, Key) {
	n := uint64(len(c.elements))
	if n > 0 {
		n = c.elements[n-1].id
	}
	return c.KeyOn(n, passphrase)
}

func (c *chain) Owned(key Key) Change {
	length := uint64(len(c.elements))
	if length == 0 {
		_ = c.Push(0, key.String(), nil, nil)
		return Change{
			N:   length,
			Gen: nil,
			Own: nil,
		}
	}
	prev := c.elements[length-1]
	nextId := prev.id + 1
	gen := createGenerator(nextId, c.token, key)
	prev.gen = gen
	var own *string = nil
	var ownId uint64
	if length >= 2 {
		third := c.elements[length-2]
		third.owner = createOwner(gen)
		own = &third.owner.data
		ownId = third.id
	}
	_ = c.Push(nextId, key.String(), nil, nil)
	return Change{
		N:     nextId,
		Gen:   &gen.data,
		GenId: prev.id,
		Own:   own,
		OwnId: ownId,
	}
}

type element struct {
	id    uint64
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

func LoadKey(hash string) Key {
	return &key{data: hash}
}

func CreateToken(data []byte) Token {
	d := sha3.Sum256(data)
	return &token{data: encodeToString(d[:])}
}

func createKey(token Token, passphrase string) Key {
	return &key{data: hash(0, token.String(), passphrase)}
}

func createGenerator(index uint64, token Token, key Key) *generator {
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
		key:   k,
		owner: createKeyOwner(k),
		gen:   createKeyGenerator(k),
	})
	return &chain{token: tok, elements: elements}
}

func CreateEmptyChain(tok string, length uint64) Chain {
	return &chain{token: &token{data: tok}, elements: make([]*element, 0), length: length}
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
