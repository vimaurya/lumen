package main

import "fmt"

var HitBuffer = make(chan Hit, 5000)

var dropCounter = 0

func Collect(hit Hit) {
	select {
	case HitBuffer <- hit:
		fmt.Print("hit sent successfully...")
	default:
		dropCounter += 1
	}
}
