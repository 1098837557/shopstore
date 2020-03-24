package main

import (
	_ "shopstore/routers"
	"github.com/astaxie/beego"
	_"shopstore/models"
)

func main() {
	beego.Run()
}

