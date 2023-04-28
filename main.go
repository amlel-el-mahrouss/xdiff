package main

func main() {
	ctx := gocsContext{}

	ctx.init("")
	ctx.track("ignore/fib.cpp")
	ctx.exit()
}
