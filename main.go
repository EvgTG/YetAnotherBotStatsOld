package main

func main() {
	app{
		cfg: NewCFG(),
		rgx: NewRegexp(),
	}.Start()
}

/*
TODO

1. топ 10 с числом проголосовавших
2. графики активности
3. опросы по темам
*/
