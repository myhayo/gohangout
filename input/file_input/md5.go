package file_input

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(str string, length int) string {
	h := md5.New()
	h.Write([]byte(str))

	return hex.EncodeToString(h.Sum(nil))[0:length]
}
