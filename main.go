package main

func main() {
	a := App{}
	a.Initialize(
		"test",
		"test",
		"home")

	a.Run(":8080")
}
