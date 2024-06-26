package crypto

import "time"

type session struct {
	hash string
	expiry   time.Time
}

// we'll use this method later to determine if the session has expired
func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func SaveHashKey() session {
	return session{
		hash: "319246ab235ecb6cb0e43017cf2b6af108e42a32737297baa1f9a559f089231b",
	}
}

func GetSessionHashKey() session {
	return session{
		hash: "319246ab235ecb6cb0e43017cf2b6af108e42a32737297baa1f9a559f089231b",
	}
}
