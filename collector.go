package main

var HitBuffer = make(chan Hit, 5000)

var dropCounter = 0

func Collect(hit Hit) {
	select {
	case HitBuffer <- hit:
	default:
		dropCounter += 1
	}
}
