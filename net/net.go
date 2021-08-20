package net

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

type Network struct {
	Name       string
	IP         string
	MACAddress string
}

type intfInfo struct {
	Name       string
	MacAddress string
	Ipv4       []string
}

func getNetworkInfo() error {
	intf, err := net.Interfaces()
	if err != nil {
		log.Fatal("get network info failed: %v", err)
		return err
	}
	var is = make([]intfInfo, len(intf))
	for i, v := range intf {
		ips, err := v.Addrs()
		if err != nil {
			log.Fatal("get network addr failed: %v", err)
			return err
		}

		//此处过滤loopback（本地回环）和isatap（isatap隧道）
		if !strings.Contains(v.Name, "lo") && !strings.Contains(v.Name, "br") {
			var network Network
			is[i].Name = v.Name
			is[i].MacAddress = v.HardwareAddr.String()
			for _, ip := range ips {
				if strings.Contains(ip.String(), ".") {
					is[i].Ipv4 = append(is[i].Ipv4, ip.String())
				}
			}
			network.Name = is[i].Name
			network.MACAddress = is[i].MacAddress
			if len(is[i].Ipv4) > 0 {
				network.IP = is[i].Ipv4[0]
			}

			fmt.Println(network.Name, network.MACAddress, network.IP)
		} else {
			fmt.Println("======", v.Name)
		}

	}

	return nil
}

// If we use bond or team technology
// to use two or more adapters as a virtual one
// use net.Interface.HardwareAddr will get the same mac address
// The following code thinking was borrowed from ethtool's source code ethtool.c
// this code will get the permanent mac address
func getPermanentMacAddr() error {
	var err error
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		fd, err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_GENERIC)
		if err != nil {
			return err
		}
	}

	defer func() {
		_ = syscall.Close(fd)
	}()
	var adapters []net.Interface
	adapters, _ = net.Interfaces()

	for _, adapter := range adapters {
		var data struct {
			cmd  uint32
			size uint32
			data [128]byte
		}

		data.cmd = 0x20
		data.size = 128
		// data.data = make([]byte, 128)

		var ifr struct {
			IfrName [16]byte
			IfrData unsafe.Pointer
		}
		copy(ifr.IfrName[:], adapter.Name)
		ifr.IfrData = unsafe.Pointer(&data)

		_, _, sysErr := syscall.RawSyscall(syscall.SYS_IOCTL,
			uintptr(fd), uintptr(0x8946), uintptr(unsafe.Pointer(&ifr)))

		if sysErr != 0 {
			fmt.Println("调用ioctl获取网口", adapter.Name, "物理MAC地址错误")
			// return errors.New("call ioctl get permanent mac address error")
			continue
		}
		// If mac address is all zero, this is a virtual adapter
		// we ignore it
		bAllZERO := true
		var i uint32 = 0
		for i = 0; i < data.size; i++ {
			if data.data[i] != 0x00 {
				bAllZERO = false
				break
			}
		}

		if bAllZERO {
			continue
		}

		if len(hex.EncodeToString(data.data[0:data.size])) > 12 {
			continue
		}

		mac := strings.ToUpper(hex.EncodeToString(data.data[0:data.size]))
		for i := 10; i > 0; i = i - 2 {
			mac = fmt.Sprintf("%s:%s", mac[:i], mac[i:])
		}

		fmt.Println(mac)
	}

	return nil
}

func IsIP(host string) bool {
	return net.ParseIP(host) != nil
}

func TelnetPort(network string, ip string, port int, timeout int) error {
	if !IsIP(ip) {
		return fmt.Errorf("IP地址不正确,ip=%s", ip)
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("端口不正确,port=%d", port)
	}

	addr := fmt.Sprintf("[%s]:%d", ip, port)
	con, err := net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("%s(%s)端口不通", addr, network)
	}
	_ = con.Close()

	return nil
}

func EqualIP(ipA, ipB string) bool {
	ip1 := net.ParseIP(ipA)
	ip2 := net.ParseIP(ipB)
	if ip1 == nil || ip2 == nil {
		return false
	}

	return ip1.Equal(ip2)
}

func IsIPv4(ip string) bool {
	if len(ip) == 0 {
		return false
	}

	if len(strings.Split(ip, ".")) != 4 {
		return false
	}

	return IsIP(ip)
}

func IsIPv6(ip string) bool {
	fmtIP, err := FormatIp(ip)
	if err != nil {
		return false
	}

	if len(fmtIP) == 0 {
		return false
	}

	if len(strings.Split(fmtIP, ":")) < 3 {
		return false
	}

	return IsIP(fmtIP)
}

// ConvertIP 当ip为IPv6时，外层加[]
func ConvertIP(ip string) string {
	trial := net.ParseIP(ip)
	if trial == nil { //not ip
		return ""
	}

	if trial.To4() == nil { //is IPv6
		return "[" + trial.String() + "]"
	}

	return trial.String()
}

// MaskToInt 将IPv4格式的子网掩码转换为整型数字
// 如 255.255.255.0 对应的整型数字为 24
// 第二个返回值string为修正后的子网掩码
func MaskToInt(netmask string) (int, string, error) {
	maskArr := strings.Split(netmask, ".")

	if len(maskArr) != 4 || !IsIP(netmask) {
		return 0, netmask, fmt.Errorf("netmask格式不正确:%v", netmask)
	}

	IPv4MaskArr := make([]byte, 4)
	IPv4MA := make([]string, 4)
	for i, value := range maskArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, "", fmt.Errorf("转换成int类型失败:[%v]", err)
		}
		divisor := 128           // 最大因子
		for i := 0; i < 8; i++ { // 非标准值自动适配
			val := 256 - divisor
			if intValue <= val && intValue != 0 {
				intValue = val
				break
			}
			divisor = divisor / 2 // 因子间隔为2倍
		}
		IPv4MA[i] = strconv.Itoa(intValue)
		IPv4MaskArr[i] = byte(intValue)
	}

	mask, _ := net.IPv4Mask(IPv4MaskArr[0], IPv4MaskArr[1], IPv4MaskArr[2], IPv4MaskArr[3]).Size()
	return mask, strings.Join(IPv4MA, "."), nil
}

// IntToMask 将掩码转换为ip格式
// 如 24 转换为 255.255.255.0
func IntToMask(mask string) (string, error) {
	nm, _ := strconv.Atoi(mask)
	if nm > 32 || nm < 0 {
		logrus.Error("mask值需介于0～32之间")
		return "", errors.New("mask值需介于0～32之间")
	}
	intMask := (0xFFFFFFFF << (32 - nm)) & 0xFFFFFFFF
	daMask := 32
	netMask := make([]string, 0, 4)
	for i := 1; i <= 4; i++ {
		tmp := intMask >> (daMask - 8) & 0xFF
		netMask = append(netMask, strconv.Itoa(tmp))
		daMask -= 8
	}

	return strings.Join(netMask, "."), nil
}

func appendIPNet(slice []net.IPNet, element net.IPNet) []net.IPNet {
	if element.IP.IsLinkLocalUnicast() { // ignore link local IPv6 address like "fe80::x"
		return slice
	}

	return append(slice, element)
}

func GetLocalIPNets() (map[string][]net.IPNet, error) {
	iFaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	returnMap := make(map[string][]net.IPNet)
	for _, iFace := range iFaces {
		if iFace.Flags&net.FlagUp == 0 { // Ignore down adapter
			continue
		}

		if iFace.Flags&net.FlagLoopback == net.FlagLoopback { // Ignore loop back adapter
			continue
		}

		addr, err := iFace.Addrs()
		if err != nil {
			continue
		}

		IPNets := make([]net.IPNet, 0)
		for _, addr := range addr {
			switch v := addr.(type) {
			case *net.IPAddr:
				IPNets = appendIPNet(IPNets, net.IPNet{IP: v.IP, Mask: v.IP.DefaultMask()})
			case *net.IPNet:
				IPNets = appendIPNet(IPNets, *v)
			}
		}
		returnMap[iFace.Name] = IPNets
	}

	return returnMap, nil
}

func GetLocalIPList() ([]string, error) {
	ipArray := make([]string, 0)
	ipMap, err := GetLocalIPNets()
	if err != nil {
		return nil, err
	}

	for _, IPNets := range ipMap {
		for _, ipNet := range IPNets {
			ipArray = append(ipArray, ipNet.IP.String())
		}
	}

	return ipArray, nil
}

func IsLocalIp(strIp string) error {
	IpStr, err := FormatIp(strIp)
	if err != nil {
		return err
	}

	ipArr, err := GetLocalIPList()
	if err != nil {
		return err
	}

	for _, value := range ipArr {
		if EqualIP(value, IpStr) {
			return nil
		}
	}

	return fmt.Errorf("非本地IP")
}

// FindBestIPNet 对比ip是与本地同网段
//find the best one(who is in same network with otherIpAddr)
func FindBestIPNet(otherIpAddr string) (ifName string, ipNetwork *net.IPNet, err error) {
	otherIP := net.ParseIP(otherIpAddr)
	if otherIP == nil {
		return "", nil, fmt.Errorf("非法IP地址(%s)", otherIpAddr)
	}

	ipMap, err := GetLocalIPNets()
	if err != nil {
		return "", nil, err
	}

	for name, IPNets := range ipMap {
		for _, ipNet := range IPNets {
			if ipNet.Contains(otherIP) {
				ifName = name
				ipNetwork = &ipNet
				err = nil
				return
			}
		}
	}

	return "", nil, fmt.Errorf("找不到与(%s)同网段的本机IP", otherIpAddr)
}

func FormatIp(ipStr string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return "", fmt.Errorf("IP地址错误")
	}
	return ip.String(), nil
}

func FormatSegmentIp(ipStr string) (string, error) {
	ipArr := strings.Split(ipStr, "-")
	var ipFormatArr = make([]string, 0)

	for _, value := range ipArr {
		str, err := FormatIp(value)
		if err != nil {
			return ipStr, err
		}
		ipFormatArr = append(ipFormatArr, str)
	}
	if len(ipArr) > 1 {
		return fmt.Sprintf("%s-%s", ipFormatArr[0], ipFormatArr[1]), nil
	}

	return fmt.Sprintf("%s", ipFormatArr[0]), nil
}

func FormatIpArray(ipArrayStr []string) ([]string, error) {
	ipArrayFormat := make([]string, 0)
	for _, v := range ipArrayStr {
		ip, err := FormatIp(v)
		if err != nil {
			return nil, err
		}
		ipArrayFormat = append(ipArrayFormat, ip)
	}
	return ipArrayFormat, nil
}

func CheckIPArray(array []string) bool {
	for _, v := range array {
		if !IsIP(v) {
			return false
		}
	}
	return true
}

func GetIPType(ip string) string {
	var str string
	if strings.Contains(ip, ":") {
		str = "IPv6"
	} else {
		str = "IPv4"
	}

	return str
}

// FormatAddrArray 用于跨域列表格式化,支持域名和ip混合数组
func FormatAddrArray(arrayStr []string) []string {
	arrayFormat := make([]string, 0)
	reg := regexp.MustCompile(`([\w-]+.){1,8}[\w-]+`) // 域名正则
	for _, v := range arrayStr {
		addr, err := FormatIp(strings.TrimSpace(v))
		if err != nil {
			if v != reg.FindString(v) || len(v) > 64 {
				break
			}
			arrayFormat = append(arrayFormat, v)
			break
		}
		arrayFormat = append(arrayFormat, addr)
	}
	return arrayFormat
}
