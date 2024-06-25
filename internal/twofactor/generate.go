package twofactor

import (
	"github.com/xlzd/gotp"
)


func GenerateTOTP(secret string) (string, int64) {
	return gotp.NewDefaultTOTP(secret).NowWithExpiration()
}

