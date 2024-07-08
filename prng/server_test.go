package prng

import (
	"slices"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestServerGenerateConfig(t *testing.T) {
	srv, err := NewServerFromSeed(DefaultSeed)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: for some reason test hangs when max >= 1E5
	max := uint64(100)

	err = srv.GenerateConfig(max)
	if err != nil {
		t.Fatal(err)
	}

	if uint64(len(srv.Config.Peers)) != max {
		t.Fatalf("expected %d, got %d", max, len(srv.Config.Peers))
	}

	publicKeys := make([]wgtypes.Key, 0)

	for _, peer := range srv.Config.Peers {

		// check public key not empty
		if len(peer.PublicKey) == 0 {
			t.Fatal("empty public key")
		}
		// check public key not already generated
		if slices.Contains(publicKeys, peer.PublicKey) {
			t.Fatalf("public key %s already in list!", peer.PublicKey)
		}
		publicKeys = append(publicKeys, peer.PublicKey)
	}
}
