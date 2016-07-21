package crypto

import (
	"crypto/rand"
	"math/big"
)

// RandBytes returns n bytes of random data.
func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// RandIntn returns a non-negative random integer in the range [0,n). It panics
// if n <= 0.
func RandIntn(n int) (int, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(r.Int64()), err
}

// RandUint64N returns a random uint64 in the range [0,n).
func RandUint64N(n uint64) (uint64, error) {
	bigN := new(big.Int)
	bigN.SetUint64(n)
	r, err := rand.Int(rand.Reader, bigN)
	return r.Uint64(), err
}

// Perm returns, as a slice of n ints, a random permutation of the integers
// [0,n).
func Perm(n int) ([]int, error) {
	m := make([]int, n)
	for i := 0; i < n; i++ {
		j, err := RandIntn(i + 1)
		if err != nil {
			return nil, err
		}

		m[i] = m[j]
		m[j] = i
	}
	return m, nil
}
