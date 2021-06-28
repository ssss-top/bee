// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package swarm

import (
	"errors"
	"math/big"
)

// Distance returns the distance between address x and address y as a (comparable) big integer using the distance metric defined in the swarm specification.
// Fails if not all addresses are of equal length.
func Distance(x, y []byte) (*big.Int, error) {
	distanceBytes, err := DistanceRaw(x, y)
	if err != nil {
		return nil, err
	}
	r := big.NewInt(0)
	r.SetBytes(distanceBytes)
	return r, nil
}

// DistanceRaw returns the distance between address x and address y in big-endian binary format using the distance metric defined in the swarm specfication.
// Fails if not all addresses are of equal length.
func DistanceRaw(x, y []byte) ([]byte, error) {
	if len(x) != len(y) {
		return nil, errors.New("address length must match")
	}
	c := make([]byte, len(x))
	for i, addr := range x {
		c[i] = addr ^ y[i]
	}
	return c, nil
}

// DistanceCmp compares x and y to a in terms of the distance metric defined in the swarm specfication.
// it returns:
// 	1 if x is closer to a than y
// 	0 if x and y are equally far apart from a (this means that x and y are the same address)
// 	-1 if x is farther from a than y
// Fails if not all addresses are of equal length.
// 判断地址a和地址x还是和地址y更相似, 根据相同的前缀
func DistanceCmp(a, x, y []byte) (int, error) {
	if len(a) != len(x) || len(a) != len(y) {
		return 0, errors.New("address length must match")
	}

	for i := range a {
		dx := x[i] ^ a[i]
		dy := y[i] ^ a[i]
		if dx == dy {
			// 如果相等
			continue
		} else if dx < dy {
			// x和a相等， y和a不相等
			return 1, nil
		}
		// 反之
		return -1, nil
	}
	return 0, nil
}
