package twofactor

import (
	"encoding/base32"

	"github.com/xlzd/gotp"
)


func GenerateTOTP(utf8string string) (string, int64) {
	secret := base32.StdEncoding.EncodeToString([]byte(utf8string))
	return gotp.NewDefaultTOTP(secret).NowWithExpiration()
}

