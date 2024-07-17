package prng

import (
	"bytes"
	"errors"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	defaultDNS string = "1.1.1.1, 1.0.0.1"
)

// Peer represents the peer config that we use to generate
// configuration for a client.
type Peer struct {
	Address       string
	PrivateKey    string
	PublicKey     string
	DNS           string
	PeerPublicKey string
	PresharedKey  string
	EndpointIP    string
}

// NewPeerFromSeed creates a new peer instance from the passed seed.
// The number `n` is the ordinal of the peer in the deterministic sequence: 1 is for the 1st peer, etc.
func NewPeerFromSeedAndNumber(seed, n uint64) (*Peer, error) {
	if n < 1 {
		return nil, errors.New("n must be >= 1")
	}

	pskBytes, err := rand256bitForNthIteration(seed, pskOffset)
	if err != nil {
		return nil, err
	}
	psk, err := wgtypes.NewKey(pskBytes)
	if err != nil {
		return nil, err
	}

	// we want to start addressing IPs after the server,
	// so the 1st peer gets 10.8.0.2/32
	ipaddr, err := GetNthIP(DefaultCIDR, int(n+1))
	if err != nil {
		return nil, err
	}

	// Now we move to a zero-index, to be able to take
	// the right index from the offset.

	n -= 1

	b, err := rand256bitForNthIteration(seed, n+rngOffset)
	if err != nil {
		return nil, err
	}
	keyPair, err := newKeyPairFromBytes(b)
	if err != nil {
		return nil, err
	}

	serverPrivateKey, err := serverPrivateKey(seed)
	if err != nil {
		return nil, err
	}

	peer := &Peer{
		Address:       ipaddr + "/32",
		DNS:           defaultDNS,
		PrivateKey:    keyPair.PrivateKey,
		PublicKey:     keyPair.PublicKey,
		PresharedKey:  psk.String(),
		PeerPublicKey: serverPrivateKey.PublicKey().String(),
	}
	return peer, nil
}

var peerConfigTemplate = `[Interface]
Address = {{ .Address }}
DNS = {{ .DNS }}
PrivateKey = {{ .PrivateKey }}

[Peer]
Endpoint = {{ .EndpointIP }}
PublicKey = {{ .PeerPublicKey }}
PresharedKey = {{ .PresharedKey }}
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`

// SerializeConfig return a byte array containing the serialization
// of this peer's config.
func (p *Peer) SerializeConfig() []byte {
	t := template.Must(template.New("peerConfig").Parse(peerConfigTemplate))
	buf := &bytes.Buffer{}
	t.Execute(buf, p)
	return buf.Bytes()
}

// KeyPair returns the keypair generated for this peer.
func (p *Peer) KeyPair() *KeyPair {
	return &KeyPair{
		PublicKey:  p.PublicKey,
		PrivateKey: p.PrivateKey,
	}
}
