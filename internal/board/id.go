package board

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

func newID(now time.Time, rng *rand.Rand) string {
	ts := strings.ToUpper(strconv.FormatInt(now.UnixMilli(), 36))
	var sfx [3]byte
	for i := range sfx {
		sfx[i] = idAlphabet[rng.Intn(len(idAlphabet))]
	}
	return "T-" + ts + "-" + strings.ToUpper(string(sfx[:]))
}
