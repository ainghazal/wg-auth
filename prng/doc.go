// Package prng implements a naive "authentication" mechanism for a wireguard server
// based on a PRNG. Concretely, we use a fast Xorshift RNG that is good enough :tm: for
// our purposes.
//
// Said purposes are ONLY to trivially generate deterministic key material to be able to
// probe a server with minimal risk of peer IP collisions. We're only concerned about
// measuring if there's external interference with the wireguard traffic, and not protecting the
// confidentiality or integrity of the established tunnels.
//
// DO NOTE that the key material generated here is *not* from a cryptographically safe source.
// You probably do not want to be running this software!!
package prng
