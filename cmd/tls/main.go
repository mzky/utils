package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/mzky/utils/tls"
)

var (
	certByte = []byte(`-----BEGIN CERTIFICATE-----
MIIFdTCCA12gAwIBAgIKEAAAAAAAAAAAADANBgkqhkiG9w0BAQsFADBmMQswCQYD
VQQGEwJDTjEQMA4GA1UECBMHQmVpSmluZzEQMA4GA1UEBxMHQmVpSmluZzERMA8G
A1UEChMIQ2VydEFpZGUxDTALBgNVBAsTBHBhd2QxETAPBgNVBAMTCHBhd2Ryb290
MCAXDTIxMDYxODA3MTE1MloYDzIwNTEwNjE3MDcxMTUyWjBmMQswCQYDVQQGEwJD
TjEQMA4GA1UECBMHQmVpSmluZzEQMA4GA1UEBxMHQmVpSmluZzERMA8GA1UEChMI
Q2VydEFpZGUxDTALBgNVBAsTBHBhd2QxETAPBgNVBAMTCHBhd2Ryb290MIICIjAN
BgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAzJ/kmWXaOakUIVYaVqrTJL1hTBzv
ctSaRWt7jbW9g/PhOoYQNKAFlOLaOQNwFPpQhseTtk+uIHolu4buX26UDWAkiYuR
bvg3FrqhGgfuV6bk2QuBkuB5szw60cKTlR3Mwigh9/r66FxnqBhOTUH9QP9UGD4C
mA8Y5gWH4RD72Br4YZhWWXr6L8rLqLQV1scw1cdJ0ezHWy0q7PKEj3DSRaQ7Vk6B
5ptXXd3cGDd/siEAKiV3+EW+i9ADkI55VjsT+1k9Tppk0YSAXpJ4dVvCXiBOv1p8
09uPUZVmBPOJCR/oeBfNOzFFcCb2IFT78aCadhfiI2kuYaVE2s9mDHMoJ+58WMaT
hIK2I8sN/L7COMFOgl19TZD+jmtf4uUV0/uzmCjtLDiL6GoZBqt6MLUuUnf+cxJZ
RtxbeLIdMmTdnDsB2Tu+wum8iRZUqlgUzW0iwpsjmR7m+1b/h4j5t6W3rCxLpnQZ
oVchpQY54biycZTMGapGmJvP9c1hZhXyx3rpnhCVBBNv/UVY+M8Q8oS96cHbYO31
A6S5x5cEMmlVUz8tVjeVaofXGKW23khT7jBxOWI78tm6VHSFILHM1Mn8p/ctfd4C
K2wrvZ7lYKeqIspVVhVR4WRcnycdB0ywXGFLSLL3OmmCeRoSCu4bhu3vft7Ezj3b
uM33fIG0CpFGMXcCAwEAAaMjMCEwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8E
BAMCAQYwDQYJKoZIhvcNAQELBQADggIBAA0+FPSlk2njwKkCtZcgA06C+bpsfuYR
KUGGHHhX3vj2i1Hty6gdrSoULRUOrptYWtBFx31+fBqkrB1610hMpLLJo7Dxs4mz
Wak74OGtjvPNyFmLsassr9gkM79wHJNnGJQfhkQnVDdK8LfoZmo8B9cHBas0nqhT
llhxgh0x1qFNyLn3aYFqRk4VZHJqKay9WCOyQapngZXtTEdj48fmJIbj9zzZ5Tko
wn5AibDWLD07veDyB8oasO+argfnLflFwABjjQscOwPSx0rjBD36YVuamR8O4fL8
eNsli8HvT924617x6Sv5joC2FANcrqtAFOH69Rpvh8wdMdvOFgb7oQJUKPgD7GJL
7DCLcYG7HtLTKpTRNclGXfNzXjFj5hMSsMQL7Q9Sl9XL5k5a63AOuLgkFGA2Ajn0
cyogsTqb93vNMHdKHHGedH5F82rYQLP/7PDQ0amajwtwKKX04HrSWUYmv4BAOWvA
fMMlVw42hzve75A16nz2jBGzDrsHeu2xa87wTnl4ixMwo7clcR+ZBuU8lAx5BRJ2
neuUh+IKFkOPAtVxSmAL0qFeZRWvqqhMDzVK8pIVvknDafreVCoHUuuT+18zIMeO
Hw51OozFVfzXoQYlanTEqCEvsT3lerGOCj37VXU2YdLztWqMf81g5KJA4mhbl8C1
pZcuKYPcZKoM
-----END CERTIFICATE-----
`)

	keyByte = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)
)

func main() {

	var ca tls.CACert
	//ca.Cert, ca.Key, _ = tls.GenerateRoot()
	//ipArray, _ := GetLocalIPList()
	//c, k, _ := ca.GenerateServer(ipArray)

	//ca.Cert, _ = tls.ReadRootCertFile("root.cer")
	//ca.Key, _ = tls.ReadPrivKeyFile("root.key")
	//c, k, _ := ca.GenerateServer([]string{"127.0.0.1"})

	ca.Cert, _ = tls.ReadRootCert(certByte)
	ca.Key, _ = tls.ReadPrivKey(keyByte)
	c, k, _ := ca.GenerateServer([]string{"192.168.136.23"})

	_ = tls.WritePEM("server.pem", c)
	_ = tls.WritePEM("server.key", k)
	//fmt.Println(generate.GenerateRoot())
	cert, _ := tls.CertificateInfo("server.pem")
	fmt.Println(cert.NotAfter.Local().Format("2006-01-02_15:04"))
}

func appendIPNet(slice []net.IPNet, element net.IPNet) []net.IPNet {
	if element.IP.IsLinkLocalUnicast() { // ignore link local IPv6 address like "fe80::x"
		return slice
	}

	return append(slice, element)
}

func GetLocalIpNets() (map[string][]net.IPNet, error) {
	iFaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	returnMap := make(map[string][]net.IPNet)
	for _, iFace := range iFaces {
		if iFace.Flags&net.FlagUp == 0 { // Ignore down adapter
			continue
		}

		addrs, err := iFace.Addrs()
		if err != nil {
			continue
		}

		ipNets := make([]net.IPNet, 0)
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPAddr:
				ipNets = appendIPNet(ipNets, net.IPNet{v.IP, v.IP.DefaultMask()})
			case *net.IPNet:
				ipNets = appendIPNet(ipNets, *v)
			}
		}
		returnMap[iFace.Name] = ipNets
	}

	return returnMap, nil
}

func GetLocalIPList() ([]string, error) {
	ipArray := make([]string, 0)
	ipMap, err := GetLocalIpNets()
	if err != nil {
		return nil, err
	}
	mapAddr := make(map[string]string) //去重
	for _, ipNets := range ipMap {
		for _, ipNet := range ipNets {
			mapAddr[ipNet.IP.String()] = ipNet.IP.String()
		}
	}

	for _, ip := range mapAddr {
		ipArray = append(ipArray, strings.TrimSpace(ip))
	}
	return ipArray, nil
}
