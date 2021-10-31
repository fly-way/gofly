package utils

import (
	"bytes"
	"encoding/binary"
	"gofly/logs"
	"math"
	"strconv"
	"strings"
)

// interface 转 string
func ToString(n interface{}) string {
	switch n := n.(type) {
	case string:
		return n
	case int:
		return strconv.Itoa(n)
	case int32:
		return strconv.Itoa(int(n))
	case int64:
		return strconv.FormatInt(n, 10)
	case uint64:
		return strconv.FormatInt(int64(n), 10)
	case float32:
		return strconv.FormatFloat(float64(n), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case []byte:
		return string(n)
	default:
		logs.Panic("unknown to string type, n:", n)
	}

	return ""
}

// 字符串拼接
func StrJoin(n ...interface{}) string {
	var builder strings.Builder

	for _, v := range n {
		builder.WriteString(ToString(v))
	}

	return builder.String()
}

// string 转 int
func StrToInt(n string) int {
	val, err := strconv.Atoi(n)
	if err != nil {
		logs.Panic("string to int err:", err, n)
	}

	return val
}

// string 转 int32
func StrToInt32(n string) int32 {
	val, err := strconv.ParseInt(n, 10, 32)
	if err != nil {
		logs.Panic("string to int32 err:", err, n)
	}

	return int32(val)
}

// string 转 int64
func StrToInt64(n string) int64 {
	val, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		logs.Panic("string to int64 err:", err, n)
	}

	return val
}

// 设置字节顺序
// 默认为高位编址
var order binary.ByteOrder = binary.BigEndian

func SetEndian(n bool) {
	if n {
		// 将高序字节存储在起始地址（高位编址）
		order = binary.BigEndian
	} else {
		// 将低序字节存储在起始地址（低位编址）
		order = binary.LittleEndian
	}
}

// interface 转 []byte
func ToByte(n interface{}) []byte {
	switch n := n.(type) {
	case string:
		return []byte(n)
	case int:
		var val = int32(n)
		dataBuf := bytes.NewBuffer([]byte{})
		if err := binary.Write(dataBuf, order, val); err != nil {
			logs.Error("ToBytes err, data:", n)
			return nil
		}
		return dataBuf.Bytes()
	case int32:
		dataBuf := bytes.NewBuffer([]byte{})
		if err := binary.Write(dataBuf, order, n); err != nil {
			logs.Error("ToBytes err, data:", n)
			return nil
		}
		return dataBuf.Bytes()
	case int64:
		bits := math.Float64bits(float64(n))
		bytes := make([]byte, 8)
		order.PutUint64(bytes, bits)
		return bytes
	case float32:
		bits := math.Float32bits(n)
		bytes := make([]byte, 4)
		order.PutUint32(bytes, bits)
		return bytes
	case float64:
		bits := math.Float64bits(n)
		bytes := make([]byte, 8)
		order.PutUint64(bytes, bits)
		return bytes
	default:
		logs.Error("ToBytes err, undefined n type:", n)
	}
	return nil
}

// 字符串反转
func StrReverse(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
