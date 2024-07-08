package prng

import "testing"

func TestGetNthIP(t *testing.T) {
	t.Run("first ip in /24", func(t *testing.T) {
		ip, err := GetNthIP("192.168.0.0/24", 1)
		if err != nil {
			t.Fatal(err)
		}
		if ip != "192.168.0.1" {
			t.Fatal("unexpected IP")
		}
	})

	t.Run("10th ip in /24", func(t *testing.T) {
		ip, err := GetNthIP("192.168.0.0/24", 10)
		if err != nil {
			t.Fatal(err)
		}
		if ip != "192.168.0.10" {
			t.Fatalf("unexpected IP: %s", ip)
		}
	})

	t.Run("254th ip in /24", func(t *testing.T) {
		ip, err := GetNthIP("192.168.0.0/24", 254)
		if err != nil {
			t.Fatal(err)
		}
		if ip != "192.168.0.254" {
			t.Fatalf("unexpected IP: %s", ip)
		}
	})
}
