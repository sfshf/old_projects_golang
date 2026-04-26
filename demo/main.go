package main

func main() {
	for i := 0; i < 1e9; i++ {
		go func() {}()
	}
}
