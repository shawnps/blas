package native

import (
	"github.com/gonum/blas"
	"github.com/gonum/internal/asm"
)

var _ blas.Float64Level3 = Implementation{}

// Dtrsm solves
//  A X = alpha B
// where X and B are m x n matrices, and A is a unit or non unit upper or lower
// triangular matrix. The result is stored in place into B. No check is made
// that A is invertible.
func (Implementation) Dtrsm(s blas.Side, ul blas.Uplo, tA blas.Transpose, d blas.Diag, m, n int, alpha float64, a []float64, lda int, b []float64, ldb int) {
	if s != blas.Left && s != blas.Right {
		panic(badSide)
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if d != blas.NonUnit && d != blas.Unit {
		panic(badDiag)
	}
	if m < 0 {
		panic(mLT0)
	}
	if n < 0 {
		panic(nLT0)
	}
	if ldb < n {
		panic(badLdB)
	}
	if s == blas.Left {
		if lda < m {
			panic(badLdA)
		}
	} else {
		if lda < n {
			panic(badLdA)
		}
	}

	if m == 0 || n == 0 {
		return
	}

	if alpha == 0 {
		for i := 0; i < m; i++ {
			btmp := b[i*ldb : i*ldb+n]
			for j := range btmp {
				btmp[j] = 0
			}
		}
		return
	}
	nonUnit := d == blas.NonUnit
	if s == blas.Left {
		if tA == blas.NoTrans {
			if ul == blas.Upper {
				for i := m - 1; i >= 0; i-- {
					btmp := b[i*ldb : i*ldb+n]
					if alpha != 1 {
						for j := range btmp {
							btmp[j] *= alpha
						}
					}
					for ka, va := range a[i*lda+i+1 : i*lda+m] {
						k := ka + i + 1
						if va != 0 {
							asm.DaxpyUnitary(-va, b[k*ldb:k*ldb+n], btmp, btmp)
						}
					}
					if nonUnit {
						tmp := 1 / a[i*lda+i]
						for j := 0; j < n; j++ {
							btmp[j] *= tmp
						}
					}
				}
				return
			}
			for i := 0; i < m; i++ {
				btmp := b[i*ldb : i*ldb+n]
				if alpha != 1 {
					for j := 0; j < n; j++ {
						btmp[j] *= alpha
					}
				}
				for k, va := range a[i*lda : i*lda+i] {
					if va != 0 {
						asm.DaxpyUnitary(-va, b[k*ldb:k*ldb+n], btmp, btmp)
					}
				}
				if nonUnit {
					tmp := 1 / a[i*lda+i]
					for j := 0; j < n; j++ {
						btmp[j] *= tmp
					}
				}
			}
			return
		}
		// Cases where a is transposed
		if ul == blas.Upper {
			for k := 0; k < m; k++ {
				btmpk := b[k*ldb : k*ldb+n]
				if nonUnit {
					tmp := 1 / a[k*lda+k]
					for j := 0; j < n; j++ {
						btmpk[j] *= tmp
					}
				}
				for ia, va := range a[k*lda+k+1 : k*lda+m] {
					i := ia + k + 1
					if va != 0 {
						btmp := b[i*ldb : i*ldb+n]
						asm.DaxpyUnitary(-va, btmpk, btmp, btmp)
					}
				}
				if alpha != 1 {
					for j := 0; j < n; j++ {
						btmpk[j] *= alpha
					}
				}
			}
			return
		}
		for k := m - 1; k >= 0; k-- {
			btmpk := b[k*ldb : k*ldb+n]
			if nonUnit {
				tmp := 1 / a[k*lda+k]
				for j := 0; j < n; j++ {
					btmpk[j] *= tmp
				}
			}
			for i, va := range a[k*lda : k*lda+k] {
				if va != 0 {
					btmp := b[i*ldb : i*ldb+n]
					asm.DaxpyUnitary(-va, btmpk, btmp, btmp)
				}
			}
			if alpha != 1 {
				for j := 0; j < n; j++ {
					btmpk[j] *= alpha
				}
			}
		}
		return
	}
	// Cases where a is to the right of X.
	if tA == blas.NoTrans {
		if ul == blas.Upper {
			for i := 0; i < m; i++ {
				btmp := b[i*ldb : i*ldb+n]
				if alpha != 1 {
					for j := 0; j < n; j++ {
						btmp[j] *= alpha
					}
				}
				for k, vb := range btmp {
					if vb != 0 {
						if btmp[k] != 0 {
							if nonUnit {
								btmp[k] /= a[k*lda+k]
							}
							btmpk := btmp[k+1 : n]
							asm.DaxpyUnitary(-btmp[k], a[k*lda+k+1:k*lda+n], btmpk, btmpk)
						}
					}
				}
			}
			return
		}
		for i := 0; i < m; i++ {
			btmp := b[i*lda : i*lda+n]
			if alpha != 1 {
				for j := 0; j < n; j++ {
					btmp[j] *= alpha
				}
			}
			for k := n - 1; k >= 0; k-- {
				if btmp[k] != 0 {
					if nonUnit {
						btmp[k] /= a[k*lda+k]
					}
					asm.DaxpyUnitary(-btmp[k], a[k*lda:k*lda+k], btmp, btmp)
				}
			}
		}
		return
	}
	// Cases where a is transposed.
	if ul == blas.Upper {
		for i := 0; i < m; i++ {
			btmp := b[i*lda : i*lda+n]
			for j := n - 1; j >= 0; j-- {
				tmp := alpha*btmp[j] - asm.DdotUnitary(a[j*lda+j+1:j*lda+n], btmp[j+1:])
				if nonUnit {
					tmp /= a[j*lda+j]
				}
				btmp[j] = tmp
			}
		}
		return
	}
	for i := 0; i < m; i++ {
		btmp := b[i*lda : i*lda+n]
		for j := 0; j < n; j++ {
			tmp := alpha*btmp[j] - asm.DdotUnitary(a[j*lda:j*lda+j], btmp)
			if nonUnit {
				tmp /= a[j*lda+j]
			}
			btmp[j] = tmp
		}
	}
}

// Dsymm performs one of
//  C = alpha * A * B + beta * C
//  C = alpha * B * A + beta * C
// where A is a symmetric matrix and B and C are m x n matrices.
func (Implementation) Dsymm(s blas.Side, ul blas.Uplo, m, n int, alpha float64, a []float64, lda int, b []float64, ldb int, beta float64, c []float64, ldc int) {
	if s != blas.Right && s != blas.Left {
		panic("goblas: bad side")
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if m < 0 {
		panic(mLT0)
	}
	if n < 0 {
		panic(nLT0)
	}
	if (lda < m && s == blas.Left) || (lda < n && s == blas.Right) {
		panic(badLdA)
	}
	if ldb < n {
		panic(badLdB)
	}
	if ldc < n {
		panic(badLdC)
	}
	if m == 0 || n == 0 {
		return
	}
	if alpha == 0 && beta == 1 {
		return
	}
	if alpha == 0 {
		if beta == 0 {
			for i := 0; i < m; i++ {
				ctmp := c[i*ldc : i*ldc+n]
				for j := range ctmp {
					ctmp[j] = 0
				}
			}
			return
		}
		for i := 0; i < m; i++ {
			ctmp := c[i*ldc : i*ldc+n]
			for j := 0; j < n; j++ {
				ctmp[j] *= beta
			}
		}
		return
	}

	isUpper := ul == blas.Upper
	if s == blas.Left {
		for i := 0; i < m; i++ {
			atmp := alpha * a[i*lda+i]
			btmp := b[i*ldb : i*ldb+n]
			ctmp := c[i*ldc : i*ldc+n]
			for j, v := range btmp {
				ctmp[j] *= beta
				ctmp[j] += atmp * v
			}

			for k := 0; k < i; k++ {
				var atmp float64
				if isUpper {
					atmp = a[k*lda+i]
				} else {
					atmp = a[i*lda+k]
				}
				atmp *= alpha
				ctmp := c[i*ldc : i*ldc+n]
				asm.DaxpyUnitary(atmp, b[k*ldb:k*ldb+n], ctmp, ctmp)
			}
			for k := i + 1; k < m; k++ {
				var atmp float64
				if isUpper {
					atmp = a[i*lda+k]
				} else {
					atmp = a[k*lda+i]
				}
				atmp *= alpha
				ctmp := c[i*ldc : i*ldc+n]
				asm.DaxpyUnitary(atmp, b[k*ldb:k*ldb+n], ctmp, ctmp)
			}
		}
		return
	}
	if isUpper {
		for i := 0; i < m; i++ {
			for j := n - 1; j >= 0; j-- {
				tmp := alpha * b[i*ldb+j]
				var tmp2 float64
				atmp := a[j*lda+j+1 : j*lda+n]
				btmp := b[i*ldb+j+1 : i*ldb+n]
				ctmp := c[i*ldc+j+1 : i*ldc+n]
				for k, v := range atmp {
					ctmp[k] += tmp * v
					tmp2 += btmp[k] * v
				}
				c[i*ldc+j] *= beta
				c[i*ldc+j] += tmp*a[j*lda+j] + alpha*tmp2
			}
		}
		return
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			tmp := alpha * b[i*ldb+j]
			var tmp2 float64
			atmp := a[j*lda : j*lda+j]
			btmp := b[i*ldb : i*ldb+j]
			ctmp := c[i*ldc : i*ldc+j]
			for k, v := range atmp {
				ctmp[k] += tmp * v
				tmp2 += btmp[k] * v
			}
			c[i*ldc+j] *= beta
			c[i*ldc+j] += tmp*a[j*lda+j] + alpha*tmp2
		}
	}
}

// Dsyrk performs the symmetric rank-k operation
//  C = alpha * A * A^T + beta*C
// where alpha and beta are scalars, C is an nxn symmetric matrix, and A
// is n x k if NonTrans, and k x n if Trans.
func (Implementation) Dsyrk(ul blas.Uplo, tA blas.Transpose, n, k int, alpha float64, a []float64, lda int, beta float64, c []float64, ldc int) {
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.Trans && tA != blas.NoTrans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if n < 0 {
		panic(nLT0)
	}
	if k < 0 {
		panic(kLT0)
	}
	if ldc < n {
		panic(badLdC)
	}
	if tA == blas.Trans {
		if lda < n {
			panic(badLdA)
		}
	} else {
		if lda < k {
			panic(badLdA)
		}
	}
	if alpha == 0 {
		if beta == 0 {
			if ul == blas.Upper {
				for i := 0; i < n; i++ {
					ctmp := c[i*ldc+i : i*ldc+n]
					for j := range ctmp {
						ctmp[j] = 0
					}
				}
				return
			}
			for i := 0; i < n; i++ {
				ctmp := c[i*ldc : i*ldc+i+1]
				for j := range ctmp {
					ctmp[j] = 0
				}
			}
			return
		}
		if ul == blas.Upper {
			for i := 0; i < n; i++ {
				ctmp := c[i*ldc+i : i*ldc+n]
				for j := range ctmp {
					ctmp[j] *= beta
				}
			}
			return
		}
		for i := 0; i < n; i++ {
			ctmp := c[i*ldc : i*ldc+i+1]
			for j := range ctmp {
				ctmp[j] *= beta
			}
		}
		return
	}
	if tA == blas.NoTrans {
		if ul == blas.Upper {
			for i := 0; i < n; i++ {
				ctmp := c[i*ldc+i : i*ldc+n]
				atmp := a[i*lda : i*lda+k]
				for jc, vc := range ctmp {
					j := jc + i
					var tmp float64
					for l, av := range a[j*lda : j*lda+k] {
						tmp += atmp[l] * av
					}
					tmp *= alpha
					tmp += vc * beta
					ctmp[jc] = tmp
				}
			}
			return
		}
		for i := 0; i < n; i++ {
			atmp := a[i*lda : i*lda+k]
			for j, vc := range c[i*ldc : i*ldc+i+1] {
				var tmp float64
				for l, va := range a[j*lda : j*lda+k] {
					tmp += atmp[l] * va
				}
				tmp *= alpha
				tmp += vc * beta
				c[i*ldc+j] = tmp
			}
		}
		return
	}
	// Cases where a is transposed.
	if ul == blas.Upper {
		for i := 0; i < n; i++ {
			ctmp := c[i*ldc+i : i*ldc+n]
			if beta != 1 {
				for j := range ctmp {
					ctmp[j] *= beta
				}
			}
			for l := 0; l < k; l++ {
				tmp := alpha * a[l*lda+i]
				if tmp != 0 {
					for j, v := range a[l*lda+i : l*lda+n] {
						ctmp[j] += tmp * v
					}
				}
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		ctmp := c[i*ldc : i*ldc+i+1]
		if beta != 0 {
			for j := range ctmp {
				ctmp[j] *= beta
			}
		}
		for l := 0; l < k; l++ {
			tmp := alpha * a[l*lda+i]
			if tmp != 0 {
				for j, v := range a[l*lda : l*lda+i+1] {
					ctmp[j] += tmp * v
				}
			}
		}
	}
}

// Dsyr2k performs a symmetric rank 2k operation
//  C = alpha * A * B^T + alpha * B * A^T + beta *C
// where C is an n x n symmetric matrix and A and B are n x k matrices if
// tA == NoTrans and k x n otherwise.
func (Implementation) Dsyr2k(ul blas.Uplo, tA blas.Transpose, n, k int, alpha float64, a []float64, lda int, b []float64, ldb int, beta float64, c []float64, ldc int) {
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.Trans && tA != blas.NoTrans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if n < 0 {
		panic(nLT0)
	}
	if k < 0 {
		panic(kLT0)
	}
	if ldc < n {
		panic(badLdC)
	}
	if tA == blas.Trans {
		if lda < n {
			panic(badLdA)
		}
		if ldb < n {
			panic(badLdB)
		}
	} else {
		if lda < k {
			panic(badLdA)
		}
		if ldb < k {
			panic(badLdB)
		}
	}
	if alpha == 0 {
		if beta == 0 {
			if ul == blas.Upper {
				for i := 0; i < n; i++ {
					ctmp := c[i*ldc+i : i*ldc+n]
					for j := range ctmp {
						ctmp[j] = 0
					}
				}
				return
			}
			for i := 0; i < n; i++ {
				ctmp := c[i*ldc : i*ldc+i+1]
				for j := range ctmp {
					ctmp[j] = 0
				}
			}
			return
		}
		if ul == blas.Upper {
			for i := 0; i < n; i++ {
				ctmp := c[i*ldc+i : i*ldc+n]
				for j := range ctmp {
					ctmp[j] *= beta
				}
			}
			return
		}
		for i := 0; i < n; i++ {
			ctmp := c[i*ldc : i*ldc+i+1]
			for j := range ctmp {
				ctmp[j] *= beta
			}
		}
		return
	}
	if tA == blas.NoTrans {
		if ul == blas.Upper {
			for i := 0; i < n; i++ {
				atmp := a[i*lda : i*lda+k]
				btmp := b[i*lda : i*lda+k]
				ctmp := c[i*ldc+i : i*ldc+n]
				for jc := range ctmp {
					j := i + jc
					var tmp1, tmp2 float64
					binner := b[j*ldb : j*ldb+k]
					for l, v := range a[j*lda : j*lda+k] {
						tmp1 += v * btmp[l]
						tmp2 += atmp[l] * binner[l]
					}
					ctmp[jc] *= beta
					ctmp[jc] += alpha * (tmp1 + tmp2)
				}
			}
			return
		}
		for i := 0; i < n; i++ {
			atmp := a[i*lda : i*lda+k]
			btmp := b[i*lda : i*lda+k]
			ctmp := c[i*ldc : i*ldc+i+1]
			for j := 0; j <= i; j++ {
				var tmp1, tmp2 float64
				binner := b[j*ldb : j*ldb+k]
				for l, v := range a[j*lda : j*lda+k] {
					tmp1 += v * btmp[l]
					tmp2 += atmp[l] * binner[l]
				}
				ctmp[j] *= beta
				ctmp[j] += alpha * (tmp1 + tmp2)
			}
		}
		return
	}
	if ul == blas.Upper {
		for i := 0; i < n; i++ {
			ctmp := c[i*ldc+i : i*ldc+n]
			if beta != 1 {
				for j := range ctmp {
					ctmp[j] *= beta
				}
			}
			for l := 0; l < k; l++ {
				tmp1 := alpha * b[l*lda+i]
				tmp2 := alpha * a[l*lda+i]
				btmp := b[l*ldb+i : l*ldb+n]
				if tmp1 != 0 || tmp2 != 0 {
					for j, v := range a[l*lda+i : l*lda+n] {
						ctmp[j] += v*tmp1 + btmp[j]*tmp2
					}
				}
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		ctmp := c[i*ldc : i*ldc+i+1]
		if beta != 1 {
			for j := range ctmp {
				ctmp[j] *= beta
			}
		}
		for l := 0; l < k; l++ {
			tmp1 := alpha * b[l*lda+i]
			tmp2 := alpha * a[l*lda+i]
			btmp := b[l*ldb : l*ldb+i+1]
			if tmp1 != 0 || tmp2 != 0 {
				for j, v := range a[l*lda : l*lda+i+1] {
					ctmp[j] += v*tmp1 + btmp[j]*tmp2
				}
			}
		}
	}
}

// Dtrmm performs a symmetric matrix multiply
//  B = alpha * A * B
// where B is an m x n matrix and A is symmetric matrix. Side and Transpose
// set the location of A relative to B and if A is transposed.
func (Implementation) Dtrmm(s blas.Side, ul blas.Uplo, tA blas.Transpose, d blas.Diag, m, n int, alpha float64, a []float64, lda int, b []float64, ldb int) {
	if s != blas.Left && s != blas.Right {
		panic(badSide)
	}
	if ul != blas.Lower && ul != blas.Upper {
		panic(badUplo)
	}
	if tA != blas.NoTrans && tA != blas.Trans && tA != blas.ConjTrans {
		panic(badTranspose)
	}
	if d != blas.NonUnit && d != blas.Unit {
		panic(badDiag)
	}
	if m < 0 {
		panic(mLT0)
	}
	if n < 0 {
		panic(nLT0)
	}
	if ldb < n {
		panic(badLdB)
	}
	if s == blas.Left {
		if lda < m {
			panic(badLdA)
		}
	} else {
		if lda < n {
			panic(badLdA)
		}
	}
	if alpha == 0 {
		for i := 0; i < m; i++ {
			btmp := b[i*ldb : i*ldb+n]
			for j := range btmp {
				btmp[j] = 0
			}
		}
		return
	}

	nonUnit := d == blas.NonUnit
	if s == blas.Left {
		if tA == blas.NoTrans {
			if ul == blas.Upper {
				for i := 0; i < m; i++ {
					tmp := alpha
					if nonUnit {
						tmp *= a[i*lda+i]
					}
					btmp := b[i*ldb : i*ldb+n]
					for j := range btmp {
						btmp[j] *= tmp
					}
					for ka, va := range a[i*lda+i+1 : i*lda+m] {
						k := ka + i + 1
						tmp := alpha * va
						if tmp != 0 {
							asm.DaxpyUnitary(tmp, b[k*ldb:k*ldb+n], btmp, btmp)
						}
					}
				}
				return
			}
			for i := m - 1; i >= 0; i-- {
				tmp := alpha
				if nonUnit {
					tmp *= a[i*lda+i]
				}
				btmp := b[i*ldb : i*ldb+n]
				for j := range btmp {
					btmp[j] *= tmp
				}
				for k, va := range a[i*lda : i*lda+i] {
					tmp := alpha * va
					if tmp != 0 {
						asm.DaxpyUnitary(tmp, b[k*ldb:k*ldb+n], btmp, btmp)
					}
				}
			}
			return
		}
		// Cases where a is transposed.
		if ul == blas.Upper {
			for k := m - 1; k >= 0; k-- {
				btmpk := b[k*ldb : k*ldb+n]
				for ia, va := range a[k*lda+k+1 : k*lda+m] {
					i := ia + k + 1
					btmp := b[i*ldb : i*ldb+n]
					tmp := alpha * va
					if tmp != 0 {
						asm.DaxpyUnitary(tmp, btmpk, btmp, btmp)
					}
				}
				tmp := alpha
				if nonUnit {
					tmp *= a[k*lda+k]
				}
				if tmp != 1 {
					for j := 0; j < n; j++ {
						btmpk[j] *= tmp
					}
				}
			}
			return
		}
		for k := 0; k < m; k++ {
			btmpk := b[k*ldb : k*ldb+n]
			for i, va := range a[k*lda : k*lda+k] {
				btmp := b[i*ldb : i*ldb+n]
				tmp := alpha * va
				if tmp != 0 {
					asm.DaxpyUnitary(tmp, btmpk, btmp, btmp)
				}
			}
			tmp := alpha
			if nonUnit {
				tmp *= a[k*lda+k]
			}
			if tmp != 1 {
				for j := 0; j < n; j++ {
					btmpk[j] *= tmp
				}
			}
		}
		return
	}
	// Cases where a is on the right
	if tA == blas.NoTrans {
		if ul == blas.Upper {
			for i := 0; i < m; i++ {
				btmp := b[i*ldb : i*ldb+n]
				for k := n - 1; k >= 0; k-- {
					tmp := alpha * btmp[k]
					if tmp != 0 {
						btmp[k] = tmp
						if nonUnit {
							btmp[k] *= a[k*lda+k]
						}
						for ja, v := range a[k*lda+k+1 : k*lda+n] {
							j := ja + k + 1
							btmp[j] += tmp * v
						}
					}
				}
			}
			return
		}
		for i := 0; i < m; i++ {
			btmp := b[i*ldb : i*ldb+n]
			for k := 0; k < n; k++ {
				tmp := alpha * btmp[k]
				if tmp != 0 {
					btmp[k] = tmp
					if nonUnit {
						btmp[k] *= a[k*lda+k]
					}
					asm.DaxpyUnitary(tmp, a[k*lda:k*lda+k], btmp, btmp)
				}
			}
		}
		return
	}
	// Cases where a is transposed.
	if ul == blas.Upper {
		for i := 0; i < m; i++ {
			btmp := b[i*lda : i*lda+n]
			for j, vb := range btmp {
				tmp := vb
				if nonUnit {
					tmp *= a[j*lda+j]
				}
				tmp += asm.DdotUnitary(a[j*lda+j+1:j*lda+n], btmp[j+1:n])
				btmp[j] = alpha * tmp
			}
		}
		return
	}
	for i := 0; i < m; i++ {
		btmp := b[i*lda : i*lda+n]
		for j := n - 1; j >= 0; j-- {
			tmp := btmp[j]
			if nonUnit {
				tmp *= a[j*lda+j]
			}
			tmp += asm.DdotUnitary(a[j*lda:j*lda+j], btmp[:j])
			btmp[j] = alpha * tmp
		}
	}
}