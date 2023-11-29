package main

import (
	"flag"
	"fmt"
	"github.com/mzky/utils/net"
	"github.com/mzky/utils/tls"
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
	ip := flag.String("ip", "127.0.0.1", fmt.Sprintf(
		"多个IP以英文逗号分割,示例：%s -ip 192.168.1.1,192.168.2.1", os.Args[0]))
	pw := flag.String("p", "123456", "p12密码")
	flag.Parse()

	var ca tls.CACert

	// 取根证书方法一
	ca.Cert, _ = tls.ReadRootCert(certByte)
	ca.Key, _ = tls.ReadPrivKey(keyByte)
	// 取根证书方法二
	ca.Cert, ca.Key, _ = tls.GenerateRoot()

	// 配置可信地址方法一：手动配置
	ipArray := []string{"127.0.0.1"}
	// 配置可信地址方法二：自动获取本地IP
	ipArray, _ = net.GetLocalIPList()

	// 产生TLS证书和密钥
	cert, key, _ := ca.GenerateServer(append(ipArray, strings.Split(*ip, ",")...))

	// 写入文件
	_ = tls.WritePEM("server.pem", cert)
	_ = tls.WritePEM("server.key", key)

	// 解析证书
	if c, err := tls.CertificateInfo("server.pem"); err != nil {
		panic(fmt.Sprintf("产生失败：%v", err))
	} else {
		fmt.Println("产生证书的有效期截至:", c.NotAfter.Local().Format("2006-01-02 15:04"))
	}

	// 产生P12证书
	p12Str, err := tls.Pkcs12Encode(cert, key, *pw)
	if err != nil {
		panic(err)
	}

	if e := os.WriteFile(*ip+".p12", []byte(p12Str), 755); e != nil {
		panic(fmt.Sprintf("写入p12证书文件失败：%v", err))
	} else {
		fmt.Println("已产生p12证书文件：", *ip+".p12")
	}
}
