package twofactor

import (
	"encoding/base32"
	"time"

	otp "github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/xlzd/gotp"
)

func GeneratePassCode(utf8string string) string {
	secret := base32.StdEncoding.EncodeToString([]byte(utf8string))

	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	})

	if err != nil {
			panic(err)
	}

	return passcode
}


func GenerateTOTP(utf8string string) (string, int64) {
	secret := base32.StdEncoding.EncodeToString([]byte(utf8string))
	return gotp.NewDefaultTOTP(secret).NowWithExpiration()
}

