package content

import "fmt"

func UnPad(src []byte) ([]byte, error) {
	length := len(src)
	padLen := int(src[length-1])

	if padLen > length {
		return nil, fmt.Errorf("padding is greater then the length: %d / %d", padLen, length)
	}

	return src[:(length - padLen)], nil
}
