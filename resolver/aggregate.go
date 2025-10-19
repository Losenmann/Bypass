package resolver

import (
	"fmt"
	"math/big"
	"net"
	"sort"
)

func ip4ToUint32(ip net.IP) uint32 {
	return uint32(ip.To4()[0])<<24 | uint32(ip.To4()[1])<<16 | uint32(ip.To4()[2])<<8 | uint32(ip.To4()[3])
}

func uint32ToIP4(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func ip6ToBigInt(ip net.IP) *big.Int {
	return new(big.Int).SetBytes(ip.To16())
}

func bigIntToIP6(n *big.Int) net.IP {
	b := n.Bytes()
	if len(b) < 16 {
		pad := make([]byte, 16-len(b))
		b = append(pad, b...)
	}
	return net.IP(b)
}

// ===== IPv4 aggregation =====
func aggregateIPv4(ips []net.IP) []string {
	if len(ips) == 0 {
		return nil
	}
	nums := make([]uint32, 0, len(ips))
	for _, ip := range ips {
		nums = append(nums, ip4ToUint32(ip))
	}
	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })

	var result []string
	i := 0
	for i < len(nums) {
		start := nums[i]
		end := start
		for i+1 < len(nums) && nums[i+1] == end+1 {
			end++
			i++
		}
		i++

		for start <= end {
			maxSize := 32
			for maxSize > 0 {
				mask := uint32(0xFFFFFFFF) << (32 - maxSize)
				if start&^mask != 0 {
					break
				}
				blockSize := uint32(1) << (32 - maxSize)
				if start+blockSize-1 > end {
					break
				}
				maxSize--
			}
			maxSize++
			result = append(result, fmt.Sprintf("%s/%d", uint32ToIP4(start), maxSize))
			start += 1 << (32 - maxSize)
		}
	}
	return result
}

// ===== IPv6 aggregation =====
func aggregateIPv6(ips []net.IP) []string {
	if len(ips) == 0 {
		return nil
	}
	nums := make([]*big.Int, 0, len(ips))
	for _, ip := range ips {
		nums = append(nums, ip6ToBigInt(ip))
	}
	sort.Slice(nums, func(i, j int) bool { return nums[i].Cmp(nums[j]) < 0 })

	var result []string
	i := 0
	for i < len(nums) {
		start := new(big.Int).Set(nums[i])
		end := new(big.Int).Set(start)
		one := big.NewInt(1)

		// объединяем подряд идущие IPv6 адреса
		for i+1 < len(nums) {
			next := new(big.Int).Set(nums[i+1])
			expected := new(big.Int).Add(end, one)
			if next.Cmp(expected) != 0 {
				break
			}
			end.Set(next)
			i++
		}
		i++

		// вычисляем CIDR блоки
		for start.Cmp(end) <= 0 {
			prefixLen := 128
			for prefixLen > 0 {
				mask := new(big.Int).Lsh(big.NewInt(1), uint(128-prefixLen))
				mask.Sub(mask, big.NewInt(1))
				inverted := new(big.Int).Not(mask)
				inverted.And(inverted, big.NewInt(0).SetBytes(make([]byte, 16)))

				if new(big.Int).And(start, mask).Cmp(big.NewInt(0)) != 0 {
					break
				}
				blockSize := new(big.Int).Lsh(big.NewInt(1), uint(128-prefixLen))
				tmp := new(big.Int).Add(start, new(big.Int).Sub(blockSize, big.NewInt(1)))
				if tmp.Cmp(end) > 0 {
					break
				}
				prefixLen--
			}
			prefixLen++
			result = append(result, fmt.Sprintf("%s/%d", bigIntToIP6(start), prefixLen))
			blockSize := new(big.Int).Lsh(big.NewInt(1), uint(128-prefixLen))
			start.Add(start, blockSize)
		}
	}
	return result
}

// ===== Универсальная функция =====
func GenCIDR(ips *[]string) []string {
	var ipv4, ipv6 []net.IP
	for _, ipStr := range *ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			ipv4 = append(ipv4, ip)
		} else {
			ipv6 = append(ipv6, ip)
		}
	}

	var result []string
	result = append(result, aggregateIPv4(ipv4)...)
	result = append(result, aggregateIPv6(ipv6)...)
	return result
}

