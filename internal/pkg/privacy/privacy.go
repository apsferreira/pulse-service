// Package privacy contém helpers de privacidade aplicados antes de persistir
// dados de telemetria. Atende LGPD Art. 7º: mantém apenas o prefixo de rede do
// IP, suficiente para análise geográfica e anti-abuso, sem identificar pessoa
// natural.
package privacy

import "net"

// AnonymizeIP retorna o IP com a porção identificadora zerada:
//   - IPv4: zera o último octeto (192.168.1.42 → 192.168.1.0)
//   - IPv6: zera os últimos 80 bits, preservando o /48 (2001:db8:abcd:1234::1 → 2001:db8:abcd::)
//
// Entradas inválidas ou vazias retornam string vazia para que o chamador não
// persista lixo. A operação é irreversível — não há sal nem chave envolvidos.
func AnonymizeIP(raw string) string {
	if raw == "" {
		return ""
	}
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		v4[3] = 0
		return v4.String()
	}
	v6 := ip.To16()
	if v6 == nil {
		return ""
	}
	for i := 6; i < 16; i++ {
		v6[i] = 0
	}
	return v6.String()
}
