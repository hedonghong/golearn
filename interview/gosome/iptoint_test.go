package gosome

import (
	"errors"
	"math"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	s1 := "192"
	s2 := []rune(s1)
	for _, v := range s2 {
		l := v - '0'
		l=l
	}
	p := '0'
	p=p
	ipstr := "192.168.1.1"
	a,_ := ipToUintSelf(ipstr)
	b, _ := ipToUint(ipstr)
	b1,_ := uintToIp(b)
	c, _ := uintToIpSelf(a)
	a=a
	b1=b1
	c=c
}

func ipToUintSelf(ip string) (uint64, error) {
	stringArr := strings.Split(ip, ".")
	unint64Arr := make([]uint64, 0, 4)
	for _, v := range stringArr {
		i := v
		p, _ := strconv.ParseUint(i, 0, 64)
		unint64Arr = append(unint64Arr, p)
	}
	// | 或 补充填位 比如 110 | 1 就是 111
	return unint64Arr[3] | unint64Arr[2] << 8 | unint64Arr[1] << 16 | unint64Arr[0] << 24, nil
}

func uintToIpSelf(ipInt uint64) (string, error) {
	i1 := ipInt >> 24//右移动24位，剩下的就是198的二进制数
	i2 := ipInt >> 16 & 0xFF//右移动16位，剩下的就是198和168的二进制数 再&上0xFF(11111111)，
	// ipInt >> 16   =  1100000001001110
	// 由于前面补0，所以是 0000000011111111
	// 由于&是1&0 0&0 = 0 1&1=1
	// 剩下的就是168的二进制
	i3 := ipInt >> 8 & 0xFF
	i4 := ipInt & 0xFF
	return strconv.FormatUint(i1, 10)+"."+strconv.FormatUint(i2, 10)+"."+strconv.FormatUint(i3, 10)+"."+strconv.FormatUint(i4, 10), nil
}

func ipToUint(ip string) (uint, error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}
	return uint(b[0]) << 24 | uint(b[1]) << 16 | uint(b[2]) << 8 | uint(b[3]), nil
	//return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
}

func uintToIp(ipInt uint) (string, error) {
	if ipInt > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(ipInt >> 24)
	ip[1] = byte(ipInt >> 16)
	ip[2] = byte(ipInt >> 8)
	ip[3] = byte(ipInt)

	return ip.String(), nil
}
