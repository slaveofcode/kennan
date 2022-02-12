package binary

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/slaveofcode/kennan/pb"
)

type Decoder struct {
	data  []byte
	index int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data:  data,
		index: 0,
	}
}

func (d *Decoder) checkEOS(length int) error {
	if d.index+length > len(d.data) {
		return io.EOF
	}
	return nil
}

func (d *Decoder) next() byte {
	val := d.data[d.index]
	d.index += 1
	return val
}

func (d *Decoder) readByte() (byte, error) {
	if err := d.checkEOS(1); err != nil {
		return 0, err
	}

	return d.next(), nil
}

// func (d *Decoder) readInt(n int, littleEndian bool) (int, error) {
// 	if err := d.checkEOS(n); err != nil {
// 		return 0, err
// 	}

// 	var val int = 0
// 	for i := 0; i < n; i++ {
// 		shift := i
// 		if !littleEndian {
// 			shift = n - 1 - i
// 		}

// 		s := shift * 8
// 		dv := d.next()
// 		cv := int64(dv) << s

// 		b := strconv.FormatInt(cv, 2)
// 		len := len(b) % 32
// 		if len == 0 {
// 			len = 32
// 		}

// 		toShift, err := strconv.ParseInt(b[:len], 2, 32)
// 		if err != nil {
// 			return 0, err
// 		}

// 		log.Println("try", dv, s, cv, toShift, b[:len], b)
// 		log.Printf("%08b", cv)
// 		log.Println("shift()", dv, cv, s)
// 		log.Println("")

// 		val |= int(toShift)
// 	}

// 	return val, nil
// }

func (d *Decoder) readInt(n int, littleEndian bool) (int, error) {
	if err := d.checkEOS(n); err != nil {
		return 0, err
	}

	var val int = 0
	for i := 0; i < n; i++ {
		shift := i
		if !littleEndian {
			shift = n - 1 - i
		}

		s := shift * 8
		dv := d.next()
		dvbig := new(big.Int)
		dvbig.SetInt64(int64(dv))

		num := new(big.Int)
		num.Lsh(dvbig, uint(s))

		b := fmt.Sprintf("%b", num)
		l := num.BitLen() % 32
		if l == 0 {
			l = 32
		}

		toShift, err := strconv.ParseInt(b[:l], 2, 32)
		if err != nil {
			return 0, err
		}

		val |= int(toShift)
	}

	return val, nil
}

func (d *Decoder) readInt20() (int, error) {
	if err := d.checkEOS(3); err != nil {
		return 0, err
	}

	return ((int(d.next()) & 15) << 16) + (int(d.next()) << 8) + int(d.next()), nil
}

func (d *Decoder) readStringFromChars(length int) (string, error) {
	if err := d.checkEOS(length); err != nil {
		return "", err
	}

	val := string(d.data[d.index : d.index+length])
	d.index += length
	return val, nil
}

func (d *Decoder) readBytes(length int) ([]byte, error) {
	if err := d.checkEOS(length); err != nil {
		return nil, err
	}

	bytes := d.data[d.index : d.index+length]
	d.index += length

	return bytes, nil
}

func (d *Decoder) unpackHex(val int) (string, error) {
	if val < 0 || val > 15 {
		return "", fmt.Errorf("invalid hex to unpack %d", val)
	}
	if val < 10 {
		return strconv.Itoa(int('0') + val), nil
	}

	return strconv.Itoa(int('A') + val - 10), nil
}

func (d *Decoder) unpackNibble(v int) (string, error) {
	if v >= 0 && v <= 9 {
		return strconv.Itoa(int('0') + v), nil
	}

	if v == 10 {
		chr := '-'
		return strconv.Itoa(int(chr)), nil
	}

	if v == 11 {
		chr := '.'
		return strconv.Itoa(int(chr)), nil
	}

	if v == 15 {
		chr := '\x00'
		return strconv.Itoa(int(chr)), nil
	}

	return "", fmt.Errorf("invalid nibble value %d", v)
}

func (d *Decoder) unpackByte(tag int, val byte) (string, error) {
	if tag == NIBBLE_8 {
		return d.unpackNibble(int(val))
	}

	if tag == HEX_8 {
		return d.unpackHex(int(val))
	}

	return "", nil
}

func (d *Decoder) readPacked8(tag int) (string, error) {
	startByte, err := d.readByte()
	if err != nil {
		return "", err
	}

	res := ""
	for i := byte(0); i < startByte&127; i++ {
		currByte, err := d.readByte()
		if err != nil {
			return "", err
		}

		val1, err := d.unpackByte(tag, (currByte&0xF0)>>4)
		if err != nil {
			return "", err
		}

		n1, err := strconv.Atoi(val1)
		if err != nil {
			return "", err
		}

		ch1 := rune(n1) // get char of ascii number
		res = res + string(ch1)

		val2, err := d.unpackByte(tag, currByte&0x0F)
		if err != nil {
			return "", err
		}

		n1, err = strconv.Atoi(val2)
		if err != nil {
			return "", err
		}

		ch1 = rune(n1) // get char of ascii number
		res = res + string(ch1)
	}

	if (startByte >> 7) != 0 {
		res = res[:len(res)-1]
	}

	return res, nil
}

func (d *Decoder) isListTag(tag int) bool {
	return tag == LIST_EMPTY || tag == LIST_8 || tag == LIST_16
}

// func (d *Decoder) readNode() (string, map[string]interface{}, interface{}, error) {
// 	b, err := d.readInt(1, false)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}

// 	lsize, err := d.readListSize(b)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}

// 	descrTag, err := d.readInt(1, false)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}

// 	if descrTag == STREAM_END {
// 		return "", nil, nil, fmt.Errorf("unexpected stream end")
// 	}

// 	descr, err := d.readString(descrTag)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}

// 	if lsize == 0 || descr == "" {
// 		return "", nil, nil, fmt.Errorf("invalid node")
// 	}

// 	attrs, err := d.readAttributes((lsize - 1) >> 1)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}

// 	log.Printf("attrs: %v\n", attrs)

// 	var content interface{}
// 	if lsize%2 == 0 {
// 		btag, err := d.readByte()
// 		if err != nil {
// 			return "", nil, nil, err
// 		}

// 		tag := int(btag)
// 		log.Println("tag:", tag, d.isListTag(tag))
// 		if d.isListTag(tag) {
// 			content, err = d.readList(tag)
// 			if err != nil {
// 				return "", nil, nil, err
// 			}
// 		} else {
// 			var decoded []byte
// 			log.Println(btag, BINARY_8)
// 			switch tag {
// 			case BINARY_8:
// 				b, err := d.readInt(1, false)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}

// 				decoded, err = d.readBytes(b)
// 				log.Println("case 1", b)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}
// 			case BINARY_20:
// 				i, err := d.readInt20()
// 				if err != nil {
// 					return "", nil, nil, err
// 				}

// 				decoded, err = d.readBytes(i)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}
// 			case BINARY_32:
// 				i, err := d.readInt(4, false)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}

// 				decoded, err = d.readBytes(i)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}
// 			default:
// 				s, err := d.readString(int(tag))
// 				if err != nil {
// 					return "", nil, nil, err
// 				}

// 				decoded = []byte(s)
// 			}

// 			if descr == "message" {
// 				s, err := d.readString(tag)
// 				if err != nil {
// 					return "", nil, nil, err
// 				}

// 				content = s
// 			} else {
// 				content = decoded
// 			}

// 		}
// 	}

// 	log.Println(">", descr, attrs, content)

// 	return descr, attrs, content, nil
// }

func (d *Decoder) readNode() (string, map[string]interface{}, interface{}, error) {
	b, err := d.readByte()
	if err != nil {
		return "", nil, nil, err
	}

	log.Println("rlz:", int(b))
	lsize, err := d.readListSize(int(b))
	if err != nil {
		return "", nil, nil, err
	}

	descrTag, err := d.readByte()
	if err != nil {
		return "", nil, nil, err
	}

	if descrTag == STREAM_END {
		return "", nil, nil, fmt.Errorf("unexpected stream end")
	}

	descr, err := d.readString(int(descrTag))
	if err != nil {
		return "", nil, nil, err
	}

	if lsize == 0 || descr == "" {
		return "", nil, nil, fmt.Errorf("invalid node")
	}

	attrs, err := d.readAttributes((lsize - 1) >> 1)
	if err != nil {
		return "", nil, nil, err
	}

	var content interface{}
	if lsize%2 == 0 {
		btag, err := d.readByte()
		if err != nil {
			return "", nil, nil, err
		}

		tag := int(btag)
		log.Println("tag", tag, d.isListTag(tag))
		if d.isListTag(tag) {
			content, err = d.readList(tag)
			if err != nil {
				return "", nil, nil, err
			}
		} else {
			var decoded []byte
			log.Println("else:", btag, BINARY_8)
			switch tag {
			case BINARY_8:
				b, err := d.readByte()
				if err != nil {
					return "", nil, nil, err
				}

				decoded, err = d.readBytes(int(b))
				log.Println("case 1", b)
				if err != nil {
					return "", nil, nil, err
				}
			case BINARY_20:
				i, err := d.readInt20()
				if err != nil {
					return "", nil, nil, err
				}

				decoded, err = d.readBytes(i)
				log.Println("case 2", i)
				if err != nil {
					return "", nil, nil, err
				}
			case BINARY_32:
				i, err := d.readInt(4, false)
				if err != nil {
					return "", nil, nil, err
				}

				decoded, err = d.readBytes(i)
				log.Println("case 3", i)
				if err != nil {
					return "", nil, nil, err
				}
			default:
				s, err := d.readString(int(tag))
				if err != nil {
					return "", nil, nil, err
				}

				decoded = []byte(s)
				log.Println("case 4", s)
			}

			if descr == "message" {
				// s, err := d.readString(tag)
				// if err != nil {
				// 	return "", nil, nil, err
				// }

				// var j map[string]interface{}
				// json.Unmarshal(decoded, &j)

				// content = j
				waMsgInfo := pb.WebMessageInfo{}
				err := proto.Unmarshal(decoded, &waMsgInfo)
				if err != nil {
					log.Println("Proto Err", err.Error())
				}

				content = waMsgInfo
			} else {
				content = decoded
			}

		}
	}

	return descr, attrs, content, nil
}

type Node struct {
	Desc    string                 `json:"desc"`
	Attrs   map[string]interface{} `json:"attrs"`
	Content interface{}            `json:"content"`
}

func (d *Decoder) readList(tag int) ([]Node, error) {
	lsize, err := d.readListSize(tag)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	for i := 0; i < lsize; i++ {
		desc, attrs, content, err := d.readNode()
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, Node{
			Desc:    desc,
			Attrs:   attrs,
			Content: content,
		})
	}

	return nodes, nil
}

func (d *Decoder) readListSize(tag int) (int, error) {
	switch tag {
	case LIST_EMPTY:
		return 0, nil
	case LIST_8:
		return d.readInt(1, false)
	case LIST_16:
		return d.readInt(2, false)
	}

	return 0, fmt.Errorf("invalid tag list size: %d", tag)
}

func (d *Decoder) getToken(index int) (string, error) {
	if index < 3 || index >= len(SINGLE_BYTE_TOKENS) {
		return "", fmt.Errorf("invalid index of token")
	}

	return SINGLE_BYTE_TOKENS[index], nil
}

func (d *Decoder) getTokenDouble(index1, index2 int) (string, error) {
	n := 256*index1 + index2
	if n < 0 || n > len(DoubleByteTokens) {
		return "", fmt.Errorf("invalid double token index")
	}

	return DoubleByteTokens[n], nil
}

// func (d *Decoder) readString(tag int) (string, error) {
// 	if tag >= 3 && tag <= 235 {
// 		return d.getToken(tag)
// 	}

// 	dicts := map[int]bool{
// 		DICTIONARY_0: true,
// 		DICTIONARY_1: true,
// 		DICTIONARY_2: true,
// 		DICTIONARY_3: true,
// 	}

// 	if dicts[tag] {
// 		b, err := d.readInt(1, false)
// 		if err != nil {
// 			return "", err
// 		}

// 		return d.getTokenDouble(tag-DICTIONARY_0, int(b))
// 	}

// 	if tag == LIST_EMPTY {
// 		return "", nil
// 	}

// 	switch tag {
// 	case BINARY_8:
// 		b, err := d.readInt(1, false)
// 		if err != nil {
// 			return "", err
// 		}

// 		s, err := d.readStringFromChars(int(b))

// 		if err != nil {
// 			return "", err
// 		}

// 		return s, nil
// 	case BINARY_20:
// 		i, err := d.readInt20()
// 		if err != nil {
// 			return "", err
// 		}

// 		s, err := d.readStringFromChars(i)
// 		if err != nil {
// 			return "", err
// 		}
// 		return s, nil
// 	case BINARY_32:
// 		i, err := d.readInt(4, false)
// 		if err != nil {
// 			return "", err
// 		}

// 		s, err := d.readStringFromChars(i)
// 		if err != nil {
// 			return "", err
// 		}

// 		return s, nil
// 	case JID_PAIR:
// 		b, err := d.readByte()
// 		if err != nil {
// 			return "", err
// 		}

// 		identity, err := d.readString(int(b))
// 		if err != nil {
// 			return "", err
// 		}

// 		b2, err := d.readByte()
// 		if err != nil {
// 			return "", err
// 		}

// 		domain, err := d.readString(int(b2))
// 		if err != nil {
// 			return "", err
// 		}

// 		return identity + "@" + domain, nil
// 	case HEX_8, NIBBLE_8:
// 		return d.readPacked8(tag)
// 	}

// 	return "", fmt.Errorf("invalid tag")
// }

func (d *Decoder) readString(tag int) (string, error) {
	if tag >= 3 && tag <= 235 {
		return d.getToken(tag)
	}

	dicts := map[int]bool{
		DICTIONARY_0: true,
		DICTIONARY_1: true,
		DICTIONARY_2: true,
		DICTIONARY_3: true,
	}

	if dicts[tag] {
		b, err := d.readByte()
		if err != nil {
			return "", err
		}

		return d.getTokenDouble(tag-DICTIONARY_0, int(b))
	}

	if tag == LIST_EMPTY {
		return "", nil
	}

	switch tag {
	case BINARY_8:
		b, err := d.readByte()
		if err != nil {
			return "", err
		}

		s, err := d.readStringFromChars(int(b))

		if err != nil {
			return "", err
		}

		return s, nil
	case BINARY_20:
		i, err := d.readInt20()
		if err != nil {
			return "", err
		}

		s, err := d.readStringFromChars(i)
		if err != nil {
			return "", err
		}
		return s, nil
	case BINARY_32:
		i, err := d.readInt(4, false)
		if err != nil {
			return "", err
		}

		s, err := d.readStringFromChars(i)
		if err != nil {
			return "", err
		}

		return s, nil
	case JID_PAIR:
		b, err := d.readByte()
		if err != nil {
			return "", err
		}

		identity, err := d.readString(int(b))
		if err != nil {
			return "", err
		}

		b2, err := d.readByte()
		if err != nil {
			return "", err
		}

		domain, err := d.readString(int(b2))
		if err != nil {
			return "", err
		}

		return identity + "@" + domain, nil
	case HEX_8, NIBBLE_8:
		return d.readPacked8(tag)
	}

	return "", fmt.Errorf("invalid tag")
}

func (d *Decoder) readAttributes(n int) (map[string]interface{}, error) {
	if n == 0 {
		return nil, nil
	}

	attrs := make(map[string]interface{})
	for i := 0; i < n; i++ {
		b, err := d.readByte()
		if err != nil {
			return nil, err
		}

		key, err := d.readString(int(b))
		if err != nil {
			return nil, err
		}

		b2, err := d.readByte()
		if err != nil {
			return nil, err
		}

		attrs[key], err = d.readString(int(b2))
		if err != nil {
			return nil, err
		}
	}

	return attrs, nil
}

func (d *Decoder) Read() (descr string, attrs map[string]interface{}, content interface{}, err error) {
	return d.readNode()
}
