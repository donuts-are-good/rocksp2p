package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
)

func bytesToBig(buffer []byte) big.Int {
	return *big.NewInt(0).SetBytes(buffer)
}

func copyBigInt(value big.Int) *big.Int {
	return big.NewInt(0).Set(&value)
}

// decode will decode a byte buffer and return big int and int
func decode(buffer []byte) (*big.Int, int) {
	result := big.NewInt(0)
	counter := 0
	var shift uint = 0
	var b uint8
	for {
		if counter >= len(buffer) {
			fmt.Println("Could not decode varint")
		}
		b = buffer[counter]
		counter++
		value := big.NewInt(0).And(big.NewInt(int64(b)), big.NewInt(0x7f))
		value = value.Lsh(value, shift)
		result = result.Add(result, value)
		shift += 7
		if b < 0x80 {
			break
		}
	}

	return result, counter
}

func encode(value big.Int) []byte {
	buffer := make([]byte, binary.MaxVarintLen64*4)
	offset := 0
	zero := big.NewInt(0)
	upper := big.NewInt(0)
	upper.SetUint64(uint64(math.Pow(2, 31)))
	upper8 := big.NewInt(0xff)
	pack := big.NewInt(0x80)
	neg := big.NewInt(-0x80)
	for value.Cmp(upper) >= 0 {
		val := copyBigInt(value)
		val = val.And(val, upper8).Or(val, pack)
		buffer[offset] = uint8(val.Uint64())
		offset++
		value = *value.Div(&value, big.NewInt(128))
	}

	for copyBigInt(value).And(&value, neg).Cmp(zero) > 0 {
		val := copyBigInt(value)
		val = val.And(val, upper8).Or(val, pack)
		buffer[offset] = uint8(val.Uint64())
		offset++
		value = *value.Rsh(&value, 7)
	}

	val := copyBigInt(value)
	val = val.Or(val, zero)
	buffer[offset] = uint8(val.Uint64())
	offset++
	return buffer[:offset]
}

// func hexToBig(value string) big.Int {
// 	var success bool

// 	result := big.NewInt(0)

// 	result, success = result.SetString(value, 16)

// 	if !success {
// 		panic("Cannot decode hexadecimal value")
// 	}

// 	return *result
// }

func ipToVarint(address net.IP) []byte {

	if address.DefaultMask() != nil {
		// found ip6
		ip := bytes.NewBuffer(address.To4()).Bytes()

		val := bytesToBig(ip)

		return encode(val)

	}

	// found ip4
	ip := bytes.NewBuffer(address).Bytes()

	val := bytesToBig(ip)

	return encode(val)
}

func varintToIP(buffer []byte) (net.IP, int) {
	val, length := decode(buffer)

	_bytes := val.Bytes()

	if val.Cmp(big.NewInt(math.MaxUint32)) <= 0 {
		return net.IPv4(_bytes[0], _bytes[1], _bytes[2], _bytes[3]), length
	}

	ip := make(net.IP, net.IPv6len)

	copy(ip, _bytes)

	return ip, length
}

// func uint8ToBig(value uint8) big.Int {
// 	return uint64ToBig(uint64(value))
// }

func uint16ToBig(value uint16) big.Int {
	return uint64ToBig(uint64(value))
}

// func uint32ToBig(value uint32) big.Int {
// 	return uint64ToBig(uint64(value))
// }

func uint64ToBig(value uint64) big.Int {
	result := big.NewInt(0)

	result.SetUint64(value)

	return *result
}
