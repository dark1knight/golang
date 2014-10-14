package main

import "fmt"

type Fraction struct {
	numerator   int
	denominator int
}

func add(a, b *Fraction) Fraction {
	return Fraction{(a.numerator * b.denominator) + (b.numerator * a.denominator),
		(a.denominator * b.denominator)}
}

// divide a by b
func divide(a, b *Fraction) Fraction {
	return Fraction{
		(a.numerator * b.denominator), (a.denominator * b.numerator)}
}

func solve(level int, f *Fraction) Fraction {
}

func main() {
	var one Fraction = Fraction{1, 1}
	var two Fraction = solve(1)

	fmt.Println(add(one, two))
}
