package sdo

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"golang.org/x/exp/slices"
	"log"
	"math"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

type SDODataType byte

// SDOOBJECT_TYPE_BOOLEAN
const (
	DATA_TYPE_BOOLEAN                     SDODataType = 0x01
	DATA_TYPE_INTEGER_8                   SDODataType = 0x02
	DATA_TYPE_INTEGER_16                  SDODataType = 0x03
	DATA_TYPE_INTEGER_32                  SDODataType = 0x04
	DATA_TYPE_UNSIGNED_8                  SDODataType = 0x05
	DATA_TYPE_UNSIGNED_16                 SDODataType = 0x06
	DATA_TYPE_UNSIGNED_32                 SDODataType = 0x07
	DATA_TYPE_REAL_32                     SDODataType = 0x08
	DATA_TYPE_VISIBLE_STRING              SDODataType = 0x09
	DATA_TYPE_OCTET_STRING                SDODataType = 0x0A
	DATA_TYPE_UNICODE_STRING              SDODataType = 0x0B
	DATA_TYPE_TIME_OF_DAY                 SDODataType = 0x0C
	DATA_TYPE_TIME_DIFFERENCE             SDODataType = 0x0D
	DATA_TYPE_DOMAIN                      SDODataType = 0x0F
	DATA_TYPE_INTEGER_24                  SDODataType = 0x10
	DATA_TYPE_REAL_64                     SDODataType = 0x11
	DATA_TYPE_INTEGER_40                  SDODataType = 0x12
	DATA_TYPE_INTEGER_48                  SDODataType = 0x13
	DATA_TYPE_INTEGER_56                  SDODataType = 0x14
	DATA_TYPE_INTEGER_64                  SDODataType = 0x15
	DATA_TYPE_UNSIGNED_24                 SDODataType = 0x16
	DATA_TYPE_UNSIGNED_40                 SDODataType = 0x18
	DATA_TYPE_UNSIGNED_48                 SDODataType = 0x19
	DATA_TYPE_UNSIGNED_56                 SDODataType = 0x1A
	DATA_TYPE_UNSIGNED_64                 SDODataType = 0x1B
	DATA_TYPE_PDO_COMMUNICATION_PARAMETER SDODataType = 0x20
	DATA_TYPE_PDO_MAPPING                 SDODataType = 0x21
	DATA_TYPE_SDO_PARAMETER               SDODataType = 0x22
	DATA_TYPE_IDENTITY                    SDODataType = 0x23
)

// IsReversed if the bytes are send in the reverse order (little endian)
func IsReversed(datatype SDODataType) bool {
	return slices.Contains([]SDODataType{
		DATA_TYPE_INTEGER_8,
		DATA_TYPE_INTEGER_16,
		DATA_TYPE_INTEGER_24,
		DATA_TYPE_INTEGER_32,
		DATA_TYPE_INTEGER_40,
		DATA_TYPE_INTEGER_48,
		DATA_TYPE_INTEGER_56,
		DATA_TYPE_INTEGER_64,
	}, datatype)
}

func DataTypeToByte(datatype SDODataType, data string) ([]byte, bool) {
	var defaultValue []byte
	var err error
	if len(data) == 0 {
		return []byte{}, true
	}

	switch datatype {
	case DATA_TYPE_INTEGER_8:
		defaultValue, err = GetIntBytes(data, 8)

	case DATA_TYPE_INTEGER_16:
		defaultValue, err = GetIntBytes(data, 16)

	case DATA_TYPE_INTEGER_24:
		defaultValue, err = GetIntBytes(data, 24)

	case DATA_TYPE_INTEGER_32:
		defaultValue, err = GetIntBytes(data, 32)

	case DATA_TYPE_INTEGER_40:
		defaultValue, err = GetIntBytes(data, 40)

	case DATA_TYPE_INTEGER_48:
		defaultValue, err = GetIntBytes(data, 48)

	case DATA_TYPE_INTEGER_56:
		defaultValue, err = GetIntBytes(data, 56)

	case DATA_TYPE_INTEGER_64:
		defaultValue, err = GetIntBytes(data, 64)

	case DATA_TYPE_UNSIGNED_8:
		defaultValue, err = GetUIntBytes(data, 8)

	case DATA_TYPE_UNSIGNED_16:
		defaultValue, err = GetUIntBytes(data, 16)

	case DATA_TYPE_UNSIGNED_24:
		defaultValue, err = GetUIntBytes(data, 24)

	case DATA_TYPE_UNSIGNED_32:
		defaultValue, err = GetUIntBytes(data, 32)

	case DATA_TYPE_UNSIGNED_40:
		defaultValue, err = GetUIntBytes(data, 40)

	case DATA_TYPE_UNSIGNED_48:
		defaultValue, err = GetUIntBytes(data, 48)

	case DATA_TYPE_UNSIGNED_56:
		defaultValue, err = GetUIntBytes(data, 56)

	case DATA_TYPE_UNSIGNED_64:
		defaultValue, err = GetUIntBytes(data, 64)

	case DATA_TYPE_BOOLEAN:
		var output int8
		data, err := strconv.ParseBool(data)
		if err != nil {
			log.Println("ERROR when converting: ", err)
			return []byte{}, true
		}

		output = 0
		if data {
			output = 1
		}

		defaultValue = append(defaultValue, byte(output))

	case DATA_TYPE_REAL_32:
		data, err1 := strconv.ParseFloat(data, 32)
		err = err1
		conversionBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(conversionBytes, math.Float32bits(float32(data)))
		defaultValue = conversionBytes

	case DATA_TYPE_REAL_64:
		data, err1 := strconv.ParseFloat(data, 64)
		err = err1
		conversionBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(conversionBytes, math.Float64bits(data))
		defaultValue = conversionBytes

	case DATA_TYPE_DOMAIN, DATA_TYPE_OCTET_STRING, DATA_TYPE_PDO_COMMUNICATION_PARAMETER, DATA_TYPE_PDO_MAPPING, DATA_TYPE_SDO_PARAMETER, DATA_TYPE_IDENTITY:
		defaultValue, err = hex.DecodeString(data)

	case DATA_TYPE_VISIBLE_STRING:
		defaultValue = []byte(data)

	case DATA_TYPE_UNICODE_STRING:
		defaultValue = []byte(data)

	case DATA_TYPE_TIME_OF_DAY:
		loc, _ := time.LoadLocation("UTC")
		offset := time.Date(1984, 1, 1, 0, 0, 0, 0, loc)
		defaultValue, err = ParseDateString(data, offset)

	case DATA_TYPE_TIME_DIFFERENCE:
		loc, _ := time.LoadLocation("UTC")
		offset := time.Date(0, 1, 1, 0, 0, 0, 0, loc)
		defaultValue, err = ParseDateString(data, offset)
	}

	if err != nil {
		log.Println("ERROR when converting: ", err)
		return []byte{}, true
	}

	return defaultValue, false
}

func ByteToDataType(datatype SDODataType, data []byte) (string, bool) {
	if data == nil || len(data) == 0 {
		return "", true
	}

	switch datatype {
	case DATA_TYPE_UNSIGNED_8, DATA_TYPE_UNSIGNED_16, DATA_TYPE_UNSIGNED_24, DATA_TYPE_UNSIGNED_32, DATA_TYPE_UNSIGNED_40, DATA_TYPE_UNSIGNED_48, DATA_TYPE_UNSIGNED_56, DATA_TYPE_UNSIGNED_64:
		return strconv.FormatUint(ParseUInt(data), 10), false

	case DATA_TYPE_INTEGER_8, DATA_TYPE_INTEGER_16, DATA_TYPE_INTEGER_24, DATA_TYPE_INTEGER_32, DATA_TYPE_INTEGER_40, DATA_TYPE_INTEGER_48, DATA_TYPE_INTEGER_56, DATA_TYPE_INTEGER_64:
		res, err := ParseInt(data)
		return strconv.FormatInt(res, 10), err != nil

	case DATA_TYPE_BOOLEAN:
		if data[0] == 1 {
			return "1", false
		} else {
			return "0", false
		}

	case DATA_TYPE_REAL_32:
		return strconv.FormatFloat(float64(math.Float32frombits(binary.LittleEndian.Uint32(data))), 'f', -1, 32), false

	case DATA_TYPE_REAL_64:
		return strconv.FormatFloat(math.Float64frombits(binary.LittleEndian.Uint64(data)), 'f', -1, 64), false

	case DATA_TYPE_DOMAIN, DATA_TYPE_OCTET_STRING, DATA_TYPE_PDO_COMMUNICATION_PARAMETER, DATA_TYPE_PDO_MAPPING, DATA_TYPE_SDO_PARAMETER, DATA_TYPE_IDENTITY:
		return hex.EncodeToString(data), false

	case DATA_TYPE_VISIBLE_STRING:
		return string(data), false

	case DATA_TYPE_UNICODE_STRING:
		start := []rune{}
		for i, w := 0, 0; i < utf8.RuneCount(data); i += w {
			runeValue, width := utf8.DecodeRune(data[i:])
			w = width

			start = append(start, runeValue)
		}

		return string(start), false

	case DATA_TYPE_TIME_OF_DAY:
		loc, _ := time.LoadLocation("UTC")
		date := time.Date(1984, 1, 1, 0, 0, 0, 0, loc)
		return ParseDate(data, date)

	case DATA_TYPE_TIME_DIFFERENCE:
		loc, _ := time.LoadLocation("UTC")
		date := time.Date(0, 1, 1, 0, 0, 0, 0, loc)
		return ParseDate(data, date)

	default:
		return "", false
	}
}

func GetUIntBytes(inputData string, bitSize int) ([]byte, error) {
	data, err := strconv.ParseUint(inputData, 10, 64)
	if err != nil {
		log.Println("ERROR when converting: ", err)
		return []byte{}, err
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, data)
	return buf[:bitSize/8], nil
}

func GetIntBytes(inputData string, bitSize int) ([]byte, error) {
	data, err := strconv.ParseInt(inputData, 10, 64)
	if err != nil {
		log.Println("ERROR when converting: ", err)
		return []byte{}, err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(data))

	// If the signed bit is set
	if b[7]&0x80 != 0 {
		// Set the signed bit at the new end position
		b[(bitSize/8)-1] |= 1 << 7
	}

	return b[:bitSize/8], nil
}

func ParseUInt(b []byte) uint64 {
	for len(b) < 8 {
		b = append(b, 0x0)
	}

	return binary.LittleEndian.Uint64(b)
}

func ParseInt(b []byte) (int64, error) {
	if len(b) > 8 {
		return 0, errors.New("value does not fit in a int64")
	}

	// Convert from little-endian to big-endian
	reverse(b)

	var n int64
	for i, v := range b {
		shift := uint((len(b) - i - 1) * 8)
		if i == 0 && v&0x80 != 0 {
			n -= 0x80 << shift
			v &= 0x7f
		}
		n += int64(v) << shift
	}
	return n, nil
}

func ParseDate(data []byte, date time.Time) (string, bool) {
	ms := binary.LittleEndian.Uint32([]byte{data[0], data[1], data[2], data[3] & 0xf0})
	days := binary.LittleEndian.Uint16([]byte{data[4], data[5]})

	date.Add(time.Duration(days*24) * time.Hour)
	date.Add(time.Duration(ms) * time.Millisecond)
	return date.Format("2006-01-02 15:04:05.000"), false
}

func ParseDateString(data string, offset time.Time) ([]byte, error) {
	date, err := time.Parse("2006-01-02 15:04:05.000", data)

	// Days / ms has 1984 as starting point
	date.Sub(offset)

	ms := date.UnixNano() / 1000
	days := uint64(ms / int64(time.Hour*24))
	ms %= int64(time.Hour * 24)

	msBytes := make([]byte, 4)
	daysBytes := make([]byte, 2)
	binary.LittleEndian.PutUint32(msBytes, uint32(ms))
	binary.LittleEndian.PutUint16(daysBytes, uint16(days))

	outputBytes := []byte{}
	outputBytes = append(outputBytes, msBytes...)
	outputBytes = append(outputBytes, daysBytes...)

	return outputBytes, err
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
