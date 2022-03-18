package common

import (
	"bytes"
	"crypto/rand"
)

// Int2Byte 将int写入指定长度的[]byte
func Int2Byte(data, len int) (ret []byte) {
	ret = make([]byte, len)
	var tmp int = 0xff
	var index uint = 0
	for index = 0; index < uint(len); index++ {
		ret[index] = byte((tmp << (index * 8) & data) >> (index * 8))
	}
	return ret
}

// Byte2Int 从[]byte中读取长度
func Byte2Int(data []byte) int {
	var ret int = 0
	var len int = len(data)
	var i uint = 0
	for i = 0; i < uint(len); i++ {
		ret = ret | (int(data[i]) << (i * 8))
	}
	return ret
}

// BytesMerger 合并[]byte
func BytesMerger(b1, b2 []byte) []byte {
	var buffer bytes.Buffer
	buffer.Write(b1)
	buffer.Write(b2)

	return buffer.Bytes()
}

// Random 随机制定长度[]byte，主要用于产生对称key
func Random(n uint) []byte {
	key := make([]byte, n)
	rand.Read(key)
	return key
}
