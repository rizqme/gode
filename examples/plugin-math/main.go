package main

import (
	"fmt"
	"math"
)
import "C"

// Plugin metadata
func Name() string { return "math" }
func Version() string { return "2.0.0" }
func Description() string { return "Mathematical operations with flexible arguments" }

// Add function with flexible arguments - works with current system
func Add(numbers ...int) (int, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("at least one number is required")
	}
	
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}

// Multiply function with optional arguments
func Multiply(a int, b int, factors ...int) int {
	result := a * b
	
	// Apply additional factors if provided
	for _, factor := range factors {
		result *= factor
	}
	
	return result
}

// Divide function with error handling
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero is not allowed")
	}
	
	return a / b, nil
}

// Power function with validation
func Power(base float64, exponent float64) (float64, error) {
	if base == 0 && exponent < 0 {
		return 0, fmt.Errorf("cannot raise zero to a negative power")
	}
	
	if base < 0 && math.Floor(exponent) != exponent {
		return 0, fmt.Errorf("cannot raise negative number to non-integer power")
	}
	
	result := math.Pow(base, exponent)
	
	if math.IsInf(result, 0) {
		return 0, fmt.Errorf("result is infinite")
	}
	
	if math.IsNaN(result) {
		return 0, fmt.Errorf("result is not a number")
	}
	
	return result, nil
}

// Fibonacci (optimized version)
func Fibonacci(n int) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("fibonacci number must be non-negative")
	}
	
	if n > 50 {
		return 0, fmt.Errorf("fibonacci number too large (max 50)")
	}
	
	if n <= 1 {
		return n, nil
	}
	
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	
	return b, nil
}

// IsPrime with enhanced algorithm
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	
	if n == 2 {
		return true
	}
	
	if n%2 == 0 {
		return false
	}
	
	// Check odd divisors up to sqrt(n)
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	
	return true
}

// Statistics function with flexible input
func Statistics(numbers ...float64) (interface{}, error) {
	if len(numbers) == 0 {
		return nil, fmt.Errorf("at least one number is required")
	}
	
	// Calculate basic statistics
	sum := 0.0
	min := numbers[0]
	max := numbers[0]
	
	for _, num := range numbers {
		sum += num
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}
	
	mean := sum / float64(len(numbers))
	
	// Calculate variance and standard deviation
	variance := 0.0
	for _, num := range numbers {
		variance += math.Pow(num-mean, 2)
	}
	variance /= float64(len(numbers))
	stdDev := math.Sqrt(variance)
	
	return map[string]interface{}{
		"count":    len(numbers),
		"sum":      sum,
		"mean":     mean,
		"min":      min,
		"max":      max,
		"range":    max - min,
		"variance": variance,
		"stdDev":   stdDev,
	}, nil
}

// GCD (Greatest Common Divisor) with flexible arguments
func GCD(numbers ...int) (int, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("at least one number is required")
	}
	
	if len(numbers) == 1 {
		return numbers[0], nil
	}
	
	// Helper function for GCD of two numbers
	gcd2 := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}
	
	result := numbers[0]
	for i := 1; i < len(numbers); i++ {
		result = gcd2(result, numbers[i])
		if result == 1 {
			break // GCD is 1, no need to continue
		}
	}
	
	return result, nil
}

// LCM (Least Common Multiple) with flexible arguments
func LCM(numbers ...int) (int, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("at least one number is required")
	}
	
	if len(numbers) == 1 {
		return numbers[0], nil
	}
	
	// Helper functions
	gcd2 := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}
	
	lcm2 := func(a, b int) int {
		return (a * b) / gcd2(a, b)
	}
	
	result := numbers[0]
	for i := 1; i < len(numbers); i++ {
		result = lcm2(result, numbers[i])
	}
	
	return result, nil
}

// FindPrimes finds all primes up to a limit
func FindPrimes(limit int) (interface{}, error) {
	if limit < 2 {
		return []int{}, nil
	}
	
	if limit > 10000 {
		return nil, fmt.Errorf("limit too large (max 10000)")
	}
	
	primes := []int{}
	
	for n := 2; n <= limit; n++ {
		if IsPrime(n) {
			primes = append(primes, n)
		}
	}
	
	return map[string]interface{}{
		"primes": primes,
		"count":  len(primes),
		"limit":  limit,
		"largest": func() int {
			if len(primes) > 0 {
				return primes[len(primes)-1]
			}
			return 0
		}(),
	}, nil
}

// Plugin interface implementation
func Initialize(rt interface{}) error {
	fmt.Println("Math plugin v2.0 initialized")
	return nil
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"add":        Add,
		"multiply":   Multiply,
		"divide":     Divide,
		"power":      Power,
		"fibonacci":  Fibonacci,
		"isPrime":    IsPrime,
		"statistics": Statistics,
		"gcd":        GCD,
		"lcm":        LCM,
		"findPrimes": FindPrimes,
	}
}

func Dispose() error {
	fmt.Println("Math plugin disposed")
	return nil
}

func main() {}