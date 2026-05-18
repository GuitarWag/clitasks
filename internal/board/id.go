package board

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

const idAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func newID(now time.Time, rng *rand.Rand) string {
	ts := strings.ToUpper(strconv.FormatInt(now.UnixMilli(), 36))
	var sfx [3]byte
	for i := range sfx {
		sfx[i] = idAlphabet[rng.IntN(len(idAlphabet))]
	}
	return "T-" + ts + "-" + string(sfx[:])
}
