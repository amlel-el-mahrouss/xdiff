package main

func main() {
	ctx := gocsContext{}
	ctx.init("")

	str := "ignore/fib.cpp"

	ctx.track(str)

	ctx.exit()
}
