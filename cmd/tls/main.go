package main

import (
	"flag"
	"fmt"
	"github.com/mzky/utils/tls"
	"net"
	"os"
	"strings"
)

var (
	certByte = []byte(`-----BEGIN CERTIFICATE-----
根证书内容
-----END CERTIFICATE-----
`)

	keyByte = []byte(`-----BEGIN RSA PRIVATE KEY-----
私钥内容
-----END RSA PRIVATE KEY-----
`)
)

func main() {
	ip := flag.String("ip", "127.0.0.1",
		fmt.Sprintf("多个IP以英文逗号分割,示例：%s -ip 192.168.1.1,192.168.2.1", os.Args[0]))
	pw := flag.String("p", "123456", "p12密码")
	flag.Parse()

	var ca tls.CACert

	ca.Cert, _ = tls.ReadRootCert(certByte)
	ca.Key, _ = tls.ReadPrivKey(keyByte)

	ca.Cert, ca.Key, _ = tls.GenerateRoot()
	ipArray := []string{"127.0.0.1"}
	c, k, _ := ca.GenerateServer(append(ipArray, strings.Split(*ip, ",")...))

	_ = tls.WritePEM("server.pem", c)
	_ = tls.WritePEM("server.key", k)
	if cert, err := tls.CertificateInfo("server.pem"); err != nil {
		panic(fmt.Sprintf("产生失败：%v", err))
	} else {
		fmt.Println("产生证书的有效期截至:", cert.NotAfter.Local().Format("2006-01-02 15:04"))
	}

	p12Str, err := tls.Pkcs12Encode(c, k, *pw)
	if err != nil {
		panic(fmt.Sprintf("pem to p12 转换证书异常: %v", err))
	}

	if e := os.WriteFile(*ip+".p12", []byte(p12Str), 755); e != nil {
		panic(fmt.Sprintln("写入p12证书文件失败：%v", err))
	} else {
		fmt.Println("已产生p12证书文件：", *ip+".p12")
	}

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
