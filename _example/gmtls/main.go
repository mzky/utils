package main

import "github.com/mzky/utils/tls"

func main() {
	if err := tls.GenKeyAndCert("sign.key", "sign.crt", 1, 0); err != nil {
		panic(err)
	}
	if err := tls.GenKeyAndCert("enc.key", "enc.crt", 2, 1); err != nil {
		panic(err)
	}
}
