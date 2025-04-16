// Package hasqchain provides an implementation of the HASQ chain building algorithm.
package hasqchain

import (
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"strings"
)

// Token represents a token in the HASQ system.
type Token interface {
	String() string
}

// Key represents a cryptographic key in the HASQ system.
type Key interface {
	String() string
}

// Change represents a change in the chain ownership.
type Change struct {
	N     uint64
	Gen   *string
	GenId uint64
	Own   *string
	OwnId uint64
}

// Chain represents a chain of ownership in the HASQ system.
type Chain interface {
	Owned(key Key) Change
	GetOwner() (uint64, Key)
	Key(passphrase string) (uint64, Key)
	KeyOn(id uint64, passphrase string) (uint64, Key)
	Validate() bool
	Push(id uint64, key string, gen *string, owner *string) error
}

// Internal implementations

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

type element struct {
	id    uint64
	key   Key
	gen   *generator
	owner *owner
}

type chain struct {
	length   uint64
	token    Token
	elements []*element
}

// String returns the string representation of a token.
func (t token) String() string {
	return t.data
}

// String returns the string representation of a key.
func (k key) String() string {
	return k.data
}

// KeyOn generates a key for the given ID and passphrase.
func (c *chain) KeyOn(id uint64, passphrase string) (uint64, Key) {
	return id, &key{data: hash(id, c.token.String(), passphrase)}
}

// GetOwner returns the owner of the chain.
func (c *chain) GetOwner() (uint64, Key) {
	l := len(c.elements)
	if l == 0 {
		return 0, nil
	}
	return c.elements[l-1].id, c.elements[l-1].key
}

// Push adds a new element to the chain.
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

// Validate checks if the chain is valid.
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

// Key generates a key for the given passphrase.
func (c *chain) Key(passphrase string) (uint64, Key) {
	n := uint64(len(c.elements))
	if n > 0 {
		n = c.elements[n-1].id
	}
	return c.KeyOn(n, passphrase)
}

// Owned establishes ownership of the chain by the given key.
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

// Helper functions

// LoadKey creates a key from a hash.
func LoadKey(hash string) Key {
	return &key{data: hash}
}

// CreateToken creates a token from data.
func CreateToken(data []byte) Token {
	d := sha3.Sum256(data)
	return &token{data: encodeToString(d[:])}
}

// CreateKey creates a key from a token and passphrase.
func CreateKey(token Token, passphrase string) Key {
	return &key{data: hash(0, token.String(), passphrase)}
}

// createGenerator creates a generator from an index, token, and key.
func createGenerator(index uint64, token Token, key Key) *generator {
	return &generator{data: hash(index, token.String(), key.String())}
}

// createOwner creates an owner from a generator.
func createOwner(generator *generator) *owner {
	return &owner{data: hash(generator.data)}
}

// createKeyOwner creates an owner from a key.
func createKeyOwner(key Key) *owner {
	return &owner{data: key.String()}
}

// createKeyGenerator creates a generator from a key.
func createKeyGenerator(key Key) *generator {
	return &generator{data: key.String()}
}

// CreateChain creates a new chain with the given token data and passphrase.
func CreateChain(tokenData []byte, passphrase string) Chain {
	tok := CreateToken(tokenData)
	k := CreateKey(tok, passphrase)
	elements := make([]*element, 0)
	elements = append(elements, &element{
		key:   k,
		owner: createKeyOwner(k),
		gen:   createKeyGenerator(k),
	})
	return &chain{token: tok, elements: elements}
}

// CreateEmptyChain creates an empty chain with the given token and length.
func CreateEmptyChain(tok string, length uint64) Chain {
	return &chain{token: &token{data: tok}, elements: make([]*element, 0), length: length}
}

// digest creates a digest from the given parameters.
func digest(params ...any) []byte {
	var h string
	for _, p := range params {
		h += fmt.Sprint(p)
	}
	d := sha3.Sum256([]byte(h))
	return d[:]
}

// hash creates a hash from the given parameters.
func hash(params ...any) string {
	return encodeToString(digest(params...))
}

// encodeToString encodes bytes to a string.
func encodeToString(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

// decodeFromString decodes a string to bytes.
func decodeFromString(data string) []byte {
	ret, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	return ret
}
