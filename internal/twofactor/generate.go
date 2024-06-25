package twofactor

import (
	"github.com/xlzd/gotp"
)

func GenerateTOTP(utf8string string) (string, int64) {
	return gotp.NewDefaultTOTP(utf8string).NowWithExpiration()
}

