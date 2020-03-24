package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"shopstore/models"
	"shopstore/tools"
	"strconv"
)

type CartController struct {
	beego.Controller
}

//添加购物车功能
func(this *CartController)HandleAddCart(){
	//获取数据
	skuid, err1 := this.GetInt("skuid")
	count, err2 := this.GetInt("count")

	//创建map来当json容器
	resp:=make(map[string]interface{})
	//发送json数据回去
	defer this.ServeJSON()


	//校验数据
	if err1!=nil || err2!=nil{
		resp["code"]=1
		resp["msg"]="传递的数据不正确"
		this.Data["json"]=resp
		beego.Info("请求数据不正确")
	}
	username := this.GetSession("userName")
	if username==nil{
		resp["code"]=1
		resp["msg"]="当前用户未登陆"
		this.Data["json"]=resp
		//this.Redirect("/login",302)
		return
	}
	//处理数据
	//购物车数据存在redis中,用hash
	o := orm.NewOrm()
	user := models.User{}
	user.Name=username.(string)
	o.Read(&user,"Name")
	conn, err3 := redis.Dial("tcp", "127.0.0.1:6379")
	if err3!=nil{
		beego.Info("redis数据库连接错误")
		return
	}

	//先获取原来的数量,再把数量加起来
	reply2, _ :=redis.Int( conn.Do("hset", "cart_"+strconv.Itoa(user.Id), skuid, count))
	conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count+reply2 )

	//获取长度
	reply, err5:= conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	//回复助手转类型
	cartCount, err6 := redis.Int(reply, err5)
	if err6!=nil{
		beego.Info("获取购物车数量失败")
	}

	resp["code"]=5
	resp["msg"]="ok"
	resp["cartCount"]=cartCount
	//指定json格式
	this.Data["json"]=resp




	//返回json数据


}

//显示购物车内容页面
func(this *CartController)ShowCart(){
	//从redis中获取商品的id与count
	username := this.GetSession("userName")
	username1,ok:=username.(string)
	if !ok{
		beego.Info("username类型转换失败")
	}

	conn, err1 := redis.Dial("tcp", "127.0.0.1:6379")
		if err1!=nil{
			beego.Info("redis拨号连接错误" )
			return
		}
	defer conn.Close()
	o := orm.NewOrm()
	user := models.User{}
	user.Name=username1
	o.Read(&user,"Name")
			//得到的是一个接口类型,要使用回复助手函数转map[string]int的切片
	result, err2 := conn.Do("hgetall", "cart_"+strconv.Itoa(user.Id))
			//使用回复助手函数
	goodMap, _ := redis.IntMap(result, err2)

	goodsInfos:=make([]map[string]interface{},len(goodMap))
		i:=0
		tatalPrice:=0
		tatalCount:=0
	for index,val:=range goodMap{
		skuid, _ := strconv.Atoi(index)
		var goodssku models.GoodsSKU
			goodssku.Id=skuid
			o.Read(&goodssku)

			temp:=make(map[string]interface{})
			temp["goods"]=goodssku
			temp["count"]=val
			goodsInfos[i]=temp
			i+=1
			//每次小计得到的值
			temp["addPrice"]=goodssku.Price*val

			//获取总价
			tatalPrice+=goodssku.Price*val
			//获取总数量
			tatalCount+=val

	}
	this.Data["tatalPrice"]=tatalPrice
	this.Data["tatalCount"]=tatalCount
	this.Data["goodsInfos"]=goodsInfos
	tools.GetUser(&this.Controller)
	this.TplName="cart.html"
}