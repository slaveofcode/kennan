package decoder

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestBinary(t *testing.T) {
	b := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	len := 63

	t.Run("Should have right length", func(t *testing.T) {
		dec := NewDecoder(b)
		i := 0
		for {
			if err := dec.checkEOS(i); err != nil {
				break
			}
			i++
		}

		if i != len {
			t.Errorf("length should %d, got %d", len, i)
		}
	})

	t.Run("Should able read bytes", func(t *testing.T) {
		dec := NewDecoder(b)
		var result []byte

		i := 0
		for {
			bVal, err := dec.readByte()
			if err != nil {
				break
			}

			result = append(result, bVal)
			i++
		}

		if !bytes.Equal(result, b) {
			t.Errorf("bytes not equals")
		}
	})

	t.Run("Should able to read partial string", func(t *testing.T) {
		dec := NewDecoder(b)
		result, err := dec.readStringFromChars(10)

		if err != nil {
			t.Errorf("got error %s", err)
		}

		if result != "abcdefghij" {
			t.Errorf("result not valid")
		}
	})

	t.Run("Should able to read partial bytes", func(t *testing.T) {
		dec := NewDecoder(b)
		b, err := dec.readBytes(10)

		if err != nil {
			t.Errorf("got error %s", err)
		}

		result := string(b)

		if result != "abcdefghij" {
			t.Errorf("result not valid")
		}
	})

	t.Run("Should able to read int 1-5", func(t *testing.T) {
		dec := NewDecoder(b)
		validResults := []int{97, 25444, 6776937, 1835954032, 1987541117}
		validEndResults := []int{98, 26213, 7105386, 1953722993, 1128415614}

		attempt := 0
		for attempt < 5 {
			v, err := dec.readInt(attempt+1, false)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if v != validResults[attempt] {
				t.Errorf("result invalid, %d != %d", v, validResults[attempt])
			}

			vend, err := dec.readInt(attempt+1, true)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if vend != validEndResults[attempt] {
				t.Errorf("result invalid, %d != %d", v, validEndResults[attempt])
			}

			attempt++
		}

	})

	t.Run("Should able to read int 6-9", func(t *testing.T) {
		dec := NewDecoder(b)
		validResults := []int{1667523942, 1887272575, 1162233676, 1448565853}
		validEndResults := []int{1785293935, 2004778364, 1548701261, 909456511}

		attempt := 5
		i := 0
		for i < 4 {
			v, err := dec.readInt(attempt+1, false)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if v != validResults[i] {
				t.Errorf("result invalid, %d != %d", v, validResults[i])
			}

			vend, err := dec.readInt(attempt+1, true)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if vend != validEndResults[i] {
				t.Errorf("result invalid, %d != %d", v, validEndResults[i])
			}

			i++
			attempt++
		}
	})

	t.Run("Should able to read int20", func(t *testing.T) {
		dec := NewDecoder(b)
		validResults := []int{90723, 288102, 485481, 682860, 880239, 29042, 226421, 423800, 621121, 148292}

		for _, cmp := range validResults {
			val, err := dec.readInt20()
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if val != cmp {
				t.Errorf("result invalid, %d != %d", val, cmp)
			}
		}
	})

	t.Run("Should able to unpackHex", func(t *testing.T) {
		dec := NewDecoder(b)
		validResults := []string{"49", "50", "51", "52", "53", "54", "55", "56", "57", "65"}

		for i := 0; i < 10; i++ {
			hex, err := dec.unpackHex(i + 1)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if hex != validResults[i] {
				t.Errorf("result invalid, %s != %s", hex, validResults[i])
			}
		}
	})

	t.Run("Should able to unpackNibble", func(t *testing.T) {
		dec := NewDecoder(b)
		testVal := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 15}
		validResults := []string{"48", "49", "50", "51", "52", "53", "54", "55", "56", "57", "45", "46", "0"}

		for i, val := range testVal {
			res, err := dec.unpackNibble(val)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if res != validResults[i] {
				t.Errorf("result invalid, %s != %s", res, validResults[i])
			}
		}
	})

	t.Run("Should able to unpackByte", func(t *testing.T) {
		dec := NewDecoder(b)
		testHex := []byte{15, 8, 5}
		testNib := []byte{15, 8, 3}
		validHexResults := []string{"70", "56", "53"}
		validNibResults := []string{"0", "56", "51"}

		for i, val := range testHex {
			res, err := dec.unpackByte(HEX_8, val)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if res != validHexResults[i] {
				t.Errorf("result invalid, %s != %s", res, validHexResults[i])
			}
		}

		for i, val := range testNib {
			res, err := dec.unpackByte(NIBBLE_8, val)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if res != validNibResults[i] {
				t.Errorf("result invalid, %s != %s", res, validHexResults[i])
			}
		}
	})

	t.Run("Should able to readPacked8 hex", func(t *testing.T) {
		dec := NewDecoder([]byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))

		validHex := []string{
			"62636465666768696A6B6C6D6E6F707172737475767778797A4142434445464748494A4B4C4D4E4F505152535455565758595A303132333435363738396162636465666768696A6B6C6D6E6F707172737475767778797A4142434445464748494A",
			"4C4D4E4F505152535455565758595A303132333435363738396162636465666768696A6B6C6D6E6F707172737475767778797A4142434445464748494A4B4C4D4E4F505152535455565758",
		}

		for i := range validHex {
			res, err := dec.readPacked8(HEX_8)
			if err != nil {
				t.Errorf("got error %s", err)
			}

			if res != validHex[i] {
				t.Errorf("result invalid, %s != %s", res, validHex[i])
			}
		}
	})

	t.Run("Should able to detect list tag", func(t *testing.T) {
		dec := NewDecoder(b)
		tags := []int{LIST_EMPTY, LIST_8, LIST_16}
		for _, v := range tags {
			if !dec.isListTag(v) {
				t.Errorf("result invalid, %d should true", v)
			}
		}
	})

	t.Run("Should able to read list size", func(t *testing.T) {
		dec := NewDecoder(b)
		tags := []int{LIST_EMPTY, LIST_8, LIST_16}
		validResults := []int{0, 97, 25187}

		for i, v := range tags {
			res, err := dec.readListSize(v)
			if err != nil {
				t.Errorf("result invalid %s", err.Error())
			}

			if res != validResults[i] {
				t.Errorf("result invalid %d", res)
			}
		}

		_, err := dec.readListSize(HEX_8)
		if err == nil {
			t.Errorf("result invalid, should return err %s", err.Error())
		}
	})

	t.Run("Should able to read string", func(t *testing.T) {
		dec := NewDecoder([]byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))

		validResults := map[int]string{
			5:        "404",
			10:       "add",
			35:       "g.us",
			JID_PAIR: "web@width",
			BINARY_8: "defghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN",
		}

		for key, val := range validResults {
			res, err := dec.readString(key)

			if err != nil {
				t.Errorf("got error %s", err.Error())
			}

			if res != val {
				t.Errorf("result invalid %s != %s", res, val)
			}
		}
	})

	t.Run("Should able to attributes", func(t *testing.T) {
		dec := NewDecoder([]byte("abcdef"))
		validJson := []byte(`{ "web": "width", "mute": "read_only", "admin": "creator" }`)

		res, err := dec.readAttributes(3)
		if err != nil {
			t.Errorf("got error %s", err.Error())
		}

		var cmp1 interface{}
		var cmp2 interface{}

		resByte, _ := json.Marshal(res)

		json.Unmarshal(resByte, &cmp1)
		json.Unmarshal(validJson, &cmp2)

		if !reflect.DeepEqual(cmp1, cmp2) {
			t.Errorf("result not valid")
		}
	})

	t.Run("Should able to getToken", func(t *testing.T) {
		dec := NewDecoder(b)
		validResults := map[int]string{
			5:  "404",
			8:  "502",
			10: "add",
			15: "battery",
			17: "body",
			21: "code",
			27: "delete",
		}

		for num, val := range validResults {
			r, err := dec.getToken(num)
			if err != nil {
				t.Errorf("got error %s", err.Error())
			}

			if r != val {
				t.Errorf("result not valid %s != %s", r, val)
			}
		}
	})

	t.Run("Should read bytes", func(t *testing.T) {
		hex, _ := hex.DecodeString("f806092f5a0a10f804f80234fc6c0a350a1b39313735323938373131313740732e77686174736170702e6e657410011a143345423030393637354537454433374141424632122b0a292a7069616e6f20726f6f6d2074696d696e6773206172653a2a0a20363a3030414d2d31323a3030414d18b3faa7f3052003f80234fc4c0a410a1b39313735323938373131313740732e77686174736170702e6e657410001a20304643454335333330463634393239433645394132434646443242433845414418bdfaa7f305c00101f80234fc930a350a1b39313735323938373131313740732e77686174736170702e6e657410011a14334542303033433742353339414644303937353312520a50536f727279206672656e2c204920636f756c646e277420756e6465727374616e6420274c69627261272e2054797065202768656c702720746f206b6e6f77207768617420616c6c20492063616e20646f18c1faa7f3052003f80234fc540a410a1b39313735323938373131313740732e77686174736170702e6e657410001a20413132333042384436423041314437393345433241453245413043313638443812090a076c69627261727918c2faa7f305")
		dec := NewDecoder(hex)

		validResults := []string{
			"+AYJL1o=",
			"ChD4BPgCNPxsCg==",
			"NQobOTE3NTI5ODcxMTE3",
			"QHMud2hhdHNhcHAubmV0EAEaFDM=",
			"RUIwMDk2NzVFN0VEMzdBQUJGMhIrCikqcA==",
			"aWFubyByb29tIHRpbWluZ3MgYXJlOioKIDY6MDBB",
			"TS0xMjowMEFNGLP6p/MFIAP4AjT8TApBChs5MTc1Mjk4NzE=",
			"MTE3QHMud2hhdHNhcHAubmV0EAAaIDBGQ0VDNTMzMEY2NDkyOUM2RQ==",
			"OUEyQ0ZGRDJCQzhFQUQYvfqn8wXAAQH4AjT8kwo1Chs5MTc1Mjk4NzExMTdA",
			"cy53aGF0c2FwcC5uZXQQARoUM0VCMDAzQzdCNTM5QUZEMDk3NTMSUgpQU29ycnkgZnI=",
			"ZW4sIEkgY291bGRuJ3QgdW5kZXJzdGFuZCAnTGlicmEnLiBUeXBlICdoZWxwJyB0byBrbm93IA==",
			"d2hhdCBhbGwgSSBjYW4gZG8Ywfqn8wUgA/gCNPxUCkEKGzkxNzUyOTg3MTExN0BzLndoYXRzYXBwLm5l",
		}

		byteNum := 5
		for _, val := range validResults {
			b, err := dec.readBytes(byteNum)
			if err != nil {
				t.Errorf("got error %s", err.Error())
			}

			r := base64.StdEncoding.EncodeToString(b)

			if val != r {
				t.Errorf("result not valid key: %d, %s != %s", byteNum, r, val)
			}

			byteNum += 5
		}

	})

	t.Run("Should read node", func(t *testing.T) {
		hex, _ := hex.DecodeString("f806092f5a0a10f804f80234fc6c0a350a1b39313735323938373131313740732e77686174736170702e6e657410011a143345423030393637354537454433374141424632122b0a292a7069616e6f20726f6f6d2074696d696e6773206172653a2a0a20363a3030414d2d31323a3030414d18b3faa7f3052003f80234fc4c0a410a1b39313735323938373131313740732e77686174736170702e6e657410001a20304643454335333330463634393239433645394132434646443242433845414418bdfaa7f305c00101f80234fc930a350a1b39313735323938373131313740732e77686174736170702e6e657410011a14334542303033433742353339414644303937353312520a50536f727279206672656e2c204920636f756c646e277420756e6465727374616e6420274c69627261272e2054797065202768656c702720746f206b6e6f77207768617420616c6c20492063616e20646f18c1faa7f3052003f80234fc540a410a1b39313735323938373131313740732e77686174736170702e6e657410001a20413132333042384436423041314437393345433241453245413043313638443812090a076c69627261727918c2faa7f305")
		dec := NewDecoder(hex)
		validJson := []byte(`["action",{"last":"true","add":"before"},[["message",null,{"key":{"remoteJid":"917529871117@s.whatsapp.net","fromMe":true,"id":"3EB009675E7ED37AABF2"},"message":{"conversation":"*piano room timings are:*\n 6:00AM-12:00AM"},"messageTimestamp":"1584004403","status":"DELIVERY_ACK"}],["message",null,{"key":{"remoteJid":"917529871117@s.whatsapp.net","fromMe":false,"id":"0FCEC5330F64929C6E9A2CFFD2BC8EAD"},"messageTimestamp":"1584004413","messageStubType":"REVOKE"}],["message",null,{"key":{"remoteJid":"917529871117@s.whatsapp.net","fromMe":true,"id":"3EB003C7B539AFD09753"},"message":{"conversation":"Sorry fren, I couldn't understand 'Libra'. Type 'help' to know what all I can do"},"messageTimestamp":"1584004417","status":"DELIVERY_ACK"}],["message",null,{"key":{"remoteJid":"917529871117@s.whatsapp.net","fromMe":false,"id":"A1230B8D6B0A1D793EC2AE2EA0C168D8"},"message":{"conversation":"library"},"messageTimestamp":"1584004418"}]]]`)

		desc, res, content, err := dec.readNode()
		if err != nil {
			t.Errorf("got error %s", err.Error())
		}

		// log.Println("========================")
		// log.Println(desc, res, content)
		// log.Println("========================")

		var cmp1 interface{}
		var cmp2 interface{}

		resByte, _ := json.Marshal(res)
		// ctnNodes := content.([]Node)
		ctnByte, _ := json.Marshal(content)

		byt, _ := json.Marshal([]interface{}{
			desc,
			res,
			content,
		})

		json.Unmarshal(byt, &cmp1)
		json.Unmarshal(validJson, &cmp2)

		log.Println("==========beg==============")
		log.Printf("%s\n", desc)
		log.Printf("%s\n", resByte)
		log.Printf("%s\n", ctnByte)
		log.Println("---")
		log.Printf("%s\n", cmp1)
		log.Println("==========mid==============")
		log.Printf("%s\n", cmp2)
		log.Println("==========end==============")

		if !reflect.DeepEqual(cmp1, cmp2) {
			t.Errorf("result not valid")
		}
	})
}
