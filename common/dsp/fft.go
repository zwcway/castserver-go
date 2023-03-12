package dsp

import "math"

var pi = make([]float64, 512)
var fi = make([]float64, 512)
var fr = make([]float64, 512)

func FFTN(pr []float64, n, rate int, logAxis bool) int {
	k := int(math.Floor(math.Log2(float64(n))))

	if len(pi) != n {
		pi = make([]float64, n)
		fi = make([]float64, n)
		fr = make([]float64, n)
	}

	var it, m, is, i, j, nv, l0 int
	var p, q, s, vr, vi, poddr, poddi float64

	for it = 0; it < n; it++ {
		m = it
		is = 0
		for i = 0; i < k; i++ {
			j = m >> 1
			is = (is << 1) + (m - (j << 1))
			m = j
		}
		fr[it] = pr[is]
		// fi[it] = pi[is]
		fi[it] = 0
	}

	pr[0] = 1.0
	pi[0] = 0.0
	p = 2 * Pi / float64(n)
	pr[1] = math.Cos(p) //将w=e^-j2pi/n用欧拉公式表示
	pi[1] = -math.Sin(p)

	for i = 2; i < n; i++ {
		p = pr[i-1] * pr[1]
		q = pi[i-1] * pi[1]
		s = (pr[i-1] + pi[i-1]) * (pr[1] + pi[1])
		pr[i] = p - q
		pi[i] = s - p - q
	}

	for it = 0; it <= n-2; it += 2 {
		vr = fr[it]
		vi = fi[it]
		fr[it] = vr + fr[it+1]
		fi[it] = vi + fi[it+1]
		fr[it+1] = vr - fr[it+1]
		fi[it+1] = vi - fi[it+1]
	}
	m = n >> 1
	nv = 2
	for l0 = k - 2; l0 >= 0; l0-- { //蝴蝶操作
		m = m >> 1
		k = nv
		nv = nv << 1
		i = (m - 1) * nv
		for it = 0; it <= i; it += nv {
			for j = 0; j < k; j++ {
				is = m * j
				p = pr[is] * fr[it+j+k]
				q = pi[is] * fi[it+j+k]
				s = pr[is] + pi[is]
				s = s * (fr[it+j+k] + fi[it+j+k])
				poddr = p - q
				poddi = s - p - q
				fr[it+j+k] = fr[it+j] - poddr
				fi[it+j+k] = fi[it+j] - poddi
				fr[it+j] = fr[it+j] + poddr
				fi[it+j] = fi[it+j] + poddi
			}
		}
	}

	n = n >> 1
	if logAxis {
		j = 0
		m = 0
		for k = 0; k < n; k = j + 1 {
			j = int(math.Pow(float64(k), 1.01))
			if j >= n {
				break
			}
			// j = k
			p = 0
			for i = k; i <= j; i++ {
				q = math.Sqrt(fr[i]*fr[i] + fi[i]*fi[i]) //幅值计算
				if q > p {
					p = q
				}
			}
			pr[m] = p
			m++
		}
		return m
	}

	for i = 0; i < n; i++ {
		pr[i] = math.Sqrt(fr[i]*fr[i] + fi[i]*fi[i]) //幅值计算
	}
	return i
}
