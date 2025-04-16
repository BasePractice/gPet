package hasqchain

import (
	"testing"
)

func TestCreateToken(t *testing.T) {
	token := CreateToken([]byte("TEST_DATA"))
	if token == nil {
		t.Fatal("Token should not be nil")
	}
	if token.String() == "" {
		t.Fatal("Token string should not be empty")
	}
	t.Logf("Created token: %s", token.String())
}

func TestCreateKey(t *testing.T) {
	token := CreateToken([]byte("TEST_DATA"))
	key := CreateKey(token, "password")
	if key == nil {
		t.Fatal("Key should not be nil")
	}
	if key.String() == "" {
		t.Fatal("Key string should not be empty")
	}
	t.Logf("Created key: %s", key.String())
}

func TestCreateChain(t *testing.T) {
	chain := CreateChain([]byte("TEST_DATA"), "password")
	if chain == nil {
		t.Fatal("Chain should not be nil")
	}

	// Validate the initial chain
	if !chain.Validate() {
		t.Fatal("Initial chain should be valid")
	}

	// Get the owner of the chain
	id, key := chain.GetOwner()
	if id != 0 {
		t.Fatalf("Initial chain owner ID should be 0, got %d", id)
	}
	if key == nil {
		t.Fatal("Initial chain owner key should not be nil")
	}
	t.Logf("Initial chain owner key: %s", key.String())
}

func TestChainOwnership(t *testing.T) {
	chain := CreateChain([]byte("TEST_DATA"), "password")

	// Add a new owner
	_, key1 := chain.Key("password1")
	change1 := chain.Owned(key1)
	if change1.N != 1 {
		t.Fatalf("First change ID should be 1, got %d", change1.N)
	}
	if change1.Gen == nil {
		t.Fatal("First change generator should not be nil")
	}

	// Add another owner
	_, key2 := chain.Key("password2")
	change2 := chain.Owned(key2)
	if change2.N != 2 {
		t.Fatalf("Second change ID should be 2, got %d", change2.N)
	}
	if change2.Gen == nil {
		t.Fatal("Second change generator should not be nil")
	}
	if change2.Own == nil {
		t.Fatal("Second change owner should not be nil")
	}

	// Validate the chain
	if !chain.Validate() {
		t.Fatal("Chain should be valid after adding owners")
	}

	// Get the current owner
	id, key := chain.GetOwner()
	if id != 2 {
		t.Fatalf("Current chain owner ID should be 2, got %d", id)
	}
	if key.String() != key2.String() {
		t.Fatalf("Current chain owner key should be %s, got %s", key2.String(), key.String())
	}
}

func TestChainValidation(t *testing.T) {
	// Create a chain
	chain := CreateChain([]byte("TEST_DATA"), "password")

	// Add multiple owners
	appendOwner(chain, "password1")
	appendOwner(chain, "password2")
	appendOwner(chain, "password3")
	appendOwner(chain, "password4")
	appendOwner(chain, "password5")

	// Validate the chain
	if !chain.Validate() {
		t.Fatal("Chain should be valid after adding multiple owners")
	}

	// Get the current owner
	id, key := chain.GetOwner()
	if id != 5 {
		t.Fatalf("Current chain owner ID should be 5, got %d", id)
	}
	if key == nil {
		t.Fatal("Current chain owner key should not be nil")
	}
}

// Helper function to append an owner to a chain
func appendOwner(ch Chain, passphrase string) {
	_, k := ch.Key(passphrase)
	ch.Owned(k)
}
