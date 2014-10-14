package newmath

import "math"

/*
 * @brief - naive primality test copied from wikipedia
 */
func IsPrime(num int64) bool {
	if num <= int64(3) {
		return num >= 2
	} else if num%int64(2) == 0 || num%int64(3) == 0 {
		return false
	}
	for i := int64(5); i <= int64(math.Sqrt(float64(num))); i += 6 {
		if num%i == 0 || num%(i+2) == 0 {
			return false
		}
	}
	return true
}
