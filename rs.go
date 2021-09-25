package main

// 参考: https://en.wikiversity.org/wiki/Reed%E2%80%93Solomon_codes_for_coders

type RS struct {
	expTable [512]int
	logTable [256]int
}

func (rs RS) GfMultNoLUT(x int, y int, prim int, field_charac_full int, carryless bool) int {
	//if field_charac_full == nil {
	//	field_charac_full = 256
	//}
	r := 0
	for y != 0 {
		if y%2 != 0 {
			if carryless == true {
				r ^= x
			} else {
				r += x
			}
		}
		y = y >> 1
		x = x << 1
		if prim > 0 && x&field_charac_full > 0 {
			x ^= prim
		}
	}
	return r
}

func (rs RS) ClMul(x, y int) int {
	z := 0
	i := 0
	for (y >> 1) > 0 {
		if y > 0 && (1<<1) > 0 {
			z ^= x << i
		}
		i += 1
	}
	return z
}

func (rs RS) ClDiv(dividend int, divisor int) int {
	dl1 := 0
	for dividend>>dl1 > 0 {
		dl1 += 1
	}
	dl2 := 0
	for divisor>>dl2 > 0 {
		dl2 += 1
	}
	if dl1 < dl2 {
		return dividend
	}

	for i := dl1 - dl2; i > -1; i-- {
		if dividend > 0 && (1<<i+dl2-1) > 0 {
			dividend ^= divisor << i
		}
	}
	return dividend
}

//func (rs RS) GfMultNoLUT(x, y int, prim int) int {
//	result := rs.ClMul(x, y)
//	if prim > 0 {
//		result = rs.ClDiv(result, prim)
//	}
//	return result
//}

func (rs *RS) InitTables(prim int) {
	x := 1
	for i := 0; i < 255; i++ {
		rs.expTable[i] = x
		rs.logTable[x] = i
		x = rs.GfMultNoLUT(x, 2, prim, 256, true)
	}
	for i := 255; i < 512; i++ {
		rs.expTable[i] = rs.expTable[i-255]
	}
}

func (rs RS) GfMul(x, y int) int {
	if x == 0 || y == 0 {
		return 0
	}
	return rs.expTable[(rs.logTable[x] + rs.logTable[y])]
}

func (rs RS) GfPow(x, power int) int {
	return rs.expTable[(rs.logTable[x]*power)%255]
}

func (rs RS) GfInverse(x int) int {
	return rs.expTable[255-rs.logTable[x]]
}

func (rs RS) GfPolyMul(p, q []int) []int {
	rLength := len(p) + len(q) - 1
	var r []int
	for i := 0; i < rLength; i++ {
		r = append(r, 0)
	}

	for j := 0; j < len(q); j++ {
		for i := 0; i < len(p); i++ {
			r[i+j] ^= rs.GfMul(p[i], q[j])
		}
	}
	return r
}

func (rs RS) RsGeneratorPoly(nsym int) []int {
	g := []int{1}
	for i := 0; i < nsym; i++ {
		g = rs.GfPolyMul(g, []int{1, rs.GfPow(2, i)})
	}
	return g
}

func (rs RS) GfPolyDiv(dividend, divisor []int) []int {
	var msgOut []int
	for i := 0; i < len(dividend)-len(divisor)-1; i++ {
		coef := dividend[i]
		if coef != 0 {
			for j := 1; j < len(divisor); j++ {
				if divisor[j] != 0 {
					msgOut[i+j] = dividend[i+j] ^ rs.GfMul(divisor[j], coef)
				}
			}
		}
	}
	return msgOut
}

func (rs RS) RsEncodeMsg(msgIn []uint, nsym int) []int {
	gen := rs.RsGeneratorPoly(nsym)
	var msgOut []int
	for i := 0; i < len(msgIn)+len(gen)-1; i++ {
		msgOut = append(msgOut, 0)
	}
	for i := 0; i < len(msgIn); i++ {
		msgOut[i] = int(msgIn[i])
	}

	for i := 0; i < len(msgIn); i++ {
		coef := msgOut[i]
		if coef != 0 {
			for j := 1; j < len(gen); j++ {
				msgOut[i+j] ^= rs.GfMul(gen[j], coef)
			}
		}
	}
	for i := 0; i < len(msgIn); i++ {
		msgOut[i] = int(msgIn[i])
	}
	return msgOut
}
