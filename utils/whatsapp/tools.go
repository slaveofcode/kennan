package whatsapp

import (
	"strconv"
	"time"
)

func GenerateLoginTag() string {
	unix := time.Now().UnixMilli()
	return strconv.FormatInt(unix, 10)
}
