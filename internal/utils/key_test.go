package utils

import (
	"testing"
)

func TestGetPublicKey(t *testing.T) {
	algo, pubkey, err := GetPublicKey("localhost:2222")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("\nAlgorithm Identifier: %s Public Key = %s\n", algo, pubkey)
}

func TestKeyGen(t *testing.T) {
	pubKey, privKey, err := MakeSSHKeyPair()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Public Key %s\n Private Key %s\n", pubKey, privKey)
}
