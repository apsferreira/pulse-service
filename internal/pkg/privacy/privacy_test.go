package privacy

import "testing"

func TestAnonymizeIP(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ipv4 público", "200.100.50.42", "200.100.50.0"},
		{"ipv4 privado", "192.168.1.42", "192.168.1.0"},
		{"ipv4 loopback", "127.0.0.1", "127.0.0.0"},
		{"ipv6 global", "2001:db8:abcd:1234:5678:90ab:cdef:0001", "2001:db8:abcd::"},
		{"ipv6 loopback", "::1", "::"},
		{"vazio retorna vazio", "", ""},
		{"inválido retorna vazio", "not-an-ip", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AnonymizeIP(tc.in)
			if got != tc.want {
				t.Errorf("AnonymizeIP(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
