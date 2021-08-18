package main

import "fmt"

type P struct {
	x,y,z float64

}

type L struct {
	p,q P
}

func main() {
	o := P{}
	line := L{o,P{x:1}}
	fmt.Println(line.q,line.p)
}
