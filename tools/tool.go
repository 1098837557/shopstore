package tools

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"shopstore/models"
	"strconv"
)

//显示用户名与退出功能的封装
func GetUser(this *beego.Controller){
	username := this.GetSession("userName")
	if username==nil{
		this.Data["username"]=""
	}else {
		//类型断言
		this.Data["username"]=username.(string)
	}
}
//封装layout
func ShowLayout(this *beego.Controller){
	o := orm.NewOrm()
	var types []models.GoodsType
	o.QueryTable("GoodsType").All(&types)
	this.Data["types"]=types
	GetUser(this)
	this.Layout="goodslayout.html"
}

//分页功能封装
func PageTool(pageCount int,pageindex int) []int{
	var pages []int
	//1.如果总页数<=5
	if pageCount<=5{
		pages=make([]int,pageCount)
		for i,_:=range pages{
			pages[i]=i+1
		}
		//2.如果传过来的页码<=3
	}else if pageindex<=3{
		//pages:=make([]int,5)
		pages=[]int{1,2,3,4,5}
		//3.如果传过来的页码pageindex>总页码数pageCount-3
	}else if  pageindex> pageCount-3{
		pages=[]int{pageCount-4,pageCount-3,pageCount-2,pageCount-1,pageCount}

	}else {
		pages=[]int{pageCount-2,pageCount-1,pageCount,pageCount+1,pageCount+2}
	}
	return pages



}

//获取购物车记录条数的封装函数
func GetCartCount(this *beego.Controller)int  {

	//1.先获取用户Id
	username := this.GetSession("userName")
	o := orm.NewOrm()
	user := models.User{}
	value, ok := username.(string)
	if !ok {
		fmt.Println("It's not ok for type string")
	}
	user.Name=value
	o.Read(&user,"Name")
	//2.从redis中获取数据
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err!=nil{
		beego.Info("redis数据库连接错误")
	}
	reply, err1 := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	//回复助手转类型
	cartCount, err2 := redis.Int(reply, err1)
	if err2!=nil{
		beego.Info("获取购物车数量失败")
	}
	return cartCount
}