package prng

import "testing"

func TestNewPeerFromSeedAndNumber(t *testing.T) {
	peer, err := NewPeerFromSeedAndNumber(DefaultSeed, 1)
	if err != nil {
		t.Fatal("not expected error")
	}
	keyPair := peer.KeyPair()

	expectedPrivateKey := "pIqAMtx1rDCVEnZbqsBEg/B8ltxeuCZgcE+eWq8VkzI="
	if keyPair.PrivateKey != expectedPrivateKey {
		t.Fatalf("expected %v, got %v", expectedPrivateKey, keyPair.PrivateKey)
	}
	expectedPublicKey := "ZnbQM42o7lmw92T1yLmyGlCXHJPxqneVPKVrLsqHHEM="
	if keyPair.PublicKey != expectedPublicKey {
		t.Fatalf("expected %v, got %v", expectedPublicKey, keyPair.PublicKey)
	}
}
