package main

import (
	"fmt"
	"os"

	"github.com/mzky/utils/tls"
)

func main() {
	fmt.Println("=== 原有功能：直接生成自签名国密证书 ===")
	// 原有功能：直接生成自签名国密证书
	if err := tls.GenKeyAndCert("sign.key", "sign.crt", 0); err != nil {
		panic(err)
	}
	if err := tls.GenKeyAndCert("enc.key", "enc.crt", 1); err != nil {
		panic(err)
	}
	fmt.Println("自签名国密证书生成成功: sign.key/sign.crt, enc.key/enc.crt")

	fmt.Println("=== 新功能：使用国密根证书生成子证书 ===")
	// 新功能1：生成国密根证书
	rootCert, err := tls.GenerateGmRoot()
	if err != nil {
		panic(fmt.Errorf("生成国密根证书失败: %v", err))
	}
	fmt.Println("国密根证书生成成功")

	// 保存根证书到文件
	if err := rootCert.SaveGmCertAndKey("root.crt", "root.key"); err != nil {
		panic(fmt.Errorf("保存根证书失败: %v", err))
	}
	fmt.Println("国密根证书已保存: root.key/root.crt")

	keyBytes, _ := os.ReadFile("root.key")
	certBytes, _ := os.ReadFile("root.crt")
	rootCert.Cert, _ = tls.ReadGmRootCert(certBytes)
	rootCert.Key, _ = tls.ReadGmPrivKey(keyBytes)

	// 新功能2：使用根证书生成国密HTTPS签名证书
	if err := rootCert.GenerateGmCert("server_sign.key", "server_sign.crt", 0); err != nil {
		panic(fmt.Errorf("生成HTTPS签名证书失败: %v", err))
	}
	fmt.Println("HTTPS签名证书生成成功: server_sign.key/server_sign.crt")

	// 新功能3：使用根证书生成国密加密证书
	if err := rootCert.GenerateGmCert("server_enc.key", "server_enc.crt", 1); err != nil {
		panic(fmt.Errorf("生成加密证书失败: %v", err))
	}
	fmt.Println("加密证书生成成功: server_enc.key/server_enc.crt")

	fmt.Println("\n所有国密证书生成完成！")
}
