package main

func main() {
	app{
		cfg: NewCFG(),
		rgx: NewRegexp(),
	}.Start()
}

/*
TODO

1. топ 30 с числом проголосовавших
2. графики активности
3. опросы по темам

Позже:
1. хештеги
2. музыка
*/
