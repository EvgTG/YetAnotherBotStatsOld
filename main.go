package main

func main() {
	app{
		cfg: NewCFG(),
		rgx: NewRegexp(),
	}.Start()
}
