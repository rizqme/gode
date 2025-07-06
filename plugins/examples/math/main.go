package main

import "C"

// Plugin metadata
func Name() string { return "math" }
func Version() string { return "1.0.0" }

// Exported functions
func Add(a, b int) int { 
	return a + b 
}

func Multiply(a, b int) int { 
	return a * b 
}

func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	
	i := 5
	for i*i <= n {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
		i += 6
	}
	return true
}

// Plugin interface implementation
func Initialize(runtime interface{}) error { 
	return nil 
}

func Exports() map[string]interface{} {
	return map[string]interface{}{
		"add":       Add,
		"multiply":  Multiply,
		"fibonacci": Fibonacci,
		"isPrime":   IsPrime,
	}
}

func Dispose() error { 
	return nil 
}

func main() {}