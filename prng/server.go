package prng

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/apex/log"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	DefaultServerAddr      string = "10.8.0.1/24"
	DefaultCIDR            string = "10.8.0.0/24"
	DefaultInterface       string = "wg0"
	DefaultEgressInterface string = "eth0"
)

var (
	DefaultListenPort int = 51820
)

type OnceWithError struct {
	once sync.Once
	err  error
}

func (o *OnceWithError) Do(f func() error) {
	o.once.Do(func() {
		o.err = f()
	})
}

func (o *OnceWithError) Err() error {
	return o.err
}

// Server is the server-side representation of the list of peers for a given WireGuard interface.
type Server struct {
	Address         string
	PrivateKey      string
	CIDR            string
	Interface       string
	EgressInterface string
	PublicAddress   string

	xorshift     *XorShift
	presharedKey [32]byte

	Config *wgtypes.Config

	maxPeers uint64
	once     *OnceWithError
}

// TODO: use a truly random key and just communicate the pubkey.
// That will be an improvement, since in that way, there's no need
// to publish the server private key (only to seed the known public key
// to fill in the peer's config)
func serverPrivateKey(seed uint64) (*wgtypes.Key, error) {
	priv, err := rand256bitForNthIteration(seed, serverOffset)
	if err != nil {
		return nil, err
	}
	privKey, err := wgtypes.NewKey(priv)
	if err != nil {
		return nil, err
	}
	return &privKey, err
}

// NewServerFromSeed creates a new server instance from the passed seed.
func NewServerFromSeed(seed uint64) (*Server, error) {
	// TODO: use a truly random key and just communicate the psk
	psk, err := rand256bitForNthIteration(seed, pskOffset)
	if err != nil {
		return nil, err
	}

	privKey, err := serverPrivateKey(seed)
	if err != nil {
		return nil, err
	}

	return &Server{
		Address:         DefaultServerAddr,
		PrivateKey:      privKey.String(),
		CIDR:            DefaultCIDR,
		Interface:       DefaultInterface,
		EgressInterface: DefaultEgressInterface,
		xorshift: &XorShift{
			state: seed,
		},
		Config: &wgtypes.Config{
			PrivateKey: privKey,
			ListenPort: &DefaultListenPort,
		},
		presharedKey: [32]byte(psk),
		maxPeers:     1,
		once:         &OnceWithError{},
	}, nil
}

// SetIPAddress sets the public IP address.
func (s *Server) SetIPAddress(addr string) error {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("expected ip:port format")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	s.PublicAddress = addr
	s.Config.ListenPort = &port
	return nil
}

// GenerateConfig sets config for a number of max peers.
// To generate config for a valid peers, we do:
// 1. Skip a fixed offset in the xorshift PRNG.
// 2. Take the next number in the sequence.
// 3. Use this number as a seed for getting 32 "random" bytes (the peer's wg private key)
// 4. Get the n-th IP in the assigned range.
func (s *Server) GenerateConfig(max uint64) error {
	s.once.Do(func() error {
		s.maxPeers = max
		n := 0

		s.xorshift.Skip(rngOffset)

		for i := rngOffset; i < max+rngOffset; i++ {

			seed := int64(s.xorshift.Next())

			b, err := rand256bitFromSeed(seed)
			if err != nil {
				log.Infof("ERROR: %v", err)
				break
			}
			keyPair, err := newKeyPairFromBytes(b)
			if err != nil {
				log.Infof("ERROR: %v", err)
				break
			}
			public, err := wgtypes.ParseKey(keyPair.PublicKey)
			if err != nil {
				log.Infof("ERROR: %v", err)
				break
			}

			// TODO: the last IP in the range should not be valid (broadcast).
			ip, err := GetNthIP(s.CIDR, n+2)
			if err != nil {
				log.Infof("ERROR: %v", err)
				return err
			}
			_, ipNet, err := net.ParseCIDR(ip + "/32")
			if err != nil {
				log.Infof("ERROR: %v", err)
				return err
			}

			peerConfig := wgtypes.PeerConfig{
				PublicKey:  public,
				AllowedIPs: []net.IPNet{*ipNet},
			}
			s.Config.Peers = append(s.Config.Peers, peerConfig)
			n += 1
		}
		return nil
	})
	if err := s.once.Err(); err != nil {
		return err
	}
	return nil
}

var srvConfigTemplate = `[Interface]
PrivateKey = {{ .PrivateKey }}
Address = {{ .Address }}
ListenPort = {{ .Config.ListenPort }}
PostUp =  iptables -t nat -A POSTROUTING -s {{ .CIDR }} -o {{ .EgressInterface }} -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport {{ .Config.ListenPort }} -j ACCEPT; iptables -A FORWARD -i {{ .Interface }} -j ACCEPT; iptables -A FORWARD -o {{ .Interface }} -j ACCEPT;

{{ range .Config.Peers }}
[Peer]
PublicKey = {{ .PublicKey }}
AllowedIPs = {{ index .AllowedIPs 0 }}
{{ end }}
`

// SerializeConfig returns a byte array with the server configuration.
func (s *Server) SerializeConfig() []byte {
	t := template.Must(template.New("serverConfig").Parse(srvConfigTemplate))
	buf := &bytes.Buffer{}
	t.Execute(buf, s)
	return buf.Bytes()
}

// A KeyPair is a pair of cryptographic keys used for WireGuard authentication.
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// rand256bitForNthIteration returns 32 "random" bytes with the given seed, and from the nth iteration
// of the PRNG.
func rand256bitForNthIteration(primarySeed uint64, n uint64) ([]byte, error) {

	// initialize xorshift with the known primary seed
	xorshift := NewXorShiftFromSeed(primarySeed)

	// skip the n first values
	xorshift.Skip(n)

	// return the first random number after initializing the stdlib RNG with the deterministic seed
	// derived from the xorshift operation.
	return rand256bitFromSeed(int64(xorshift.Next()))
}

func rand256bitFromSeed(seed int64) ([]byte, error) {
	// initialize stdlib RNG with the passed seed
	r := rand.New(rand.NewSource(seed))

	// wireguard keys are 32 byte arrays
	b := make([]byte, 32)
	_, err := r.Read(b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func newKeyPairFromBytes(b []byte) (*KeyPair, error) {
	privateKey, err := wgtypes.NewKey(b)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PrivateKey: privateKey.String(),
		PublicKey:  privateKey.PublicKey().String(),
	}, nil
}
