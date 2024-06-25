package twofactor

import (
	"github.com/xlzd/gotp"
)


func GenerateTOTP(utf8string string) (string, int64) {
	secret := utf8string
	return gotp.NewDefaultTOTP(secret).NowWithExpiration()
}

