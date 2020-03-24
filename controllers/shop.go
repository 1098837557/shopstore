package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"math"
	"shopstore/models"
	"strconv"
	"shopstore/tools"
)

type ShopController struct {
	beego.Controller
}


//展示首页
func (this *ShopController)ShowIndex(){
	o := orm.NewOrm()
	//获取类型数据
	var goodsType []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsType)
	this.Data["goodsType"]=goodsType


	//2.获取轮播图数据
	var indexGoodsBanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanner)
	this.Data["IndexGoodsBanner"]=indexGoodsBanner

	//3.获取促销商品
	var indexPromotionBanner []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexPromotionBanner)
	this.Data["indexPromotionBanner"]=indexPromotionBanner

	//4.获取首页展示商品
	   //创建一个map的数组,string为键 ,interface类型为值
	goods:=make([]map[string]interface{},len(goodsType))
	for key ,value:=range goodsType{
		//获取类型的首页展示商品
		  //创建一个temp变量, map类型,string为键 ,interface为值
		 temp:=make( map[string]interface{})
		 //让temp的键为"type",值为循环遍历goodsType得到的每一条数据
		temp["type"]=value
			//把每次获得到的temp数据,赋值给goods数组
		goods[key]=temp
	}
	for _,value:=range goods{
		//获取文字商品数据
		var textGoods []models.IndexTypeGoodsBanner
		//查询IndexTypeGoodsBanner表,关联GoodsType表与GoodsSKU表,排序Index,过滤GoodsType表中的
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",value["type"]).Filter("DisplayType",0).All(&textGoods)

		//获取图片商品文字
		var imageGoods []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",value["type"]).Filter("DisplayType",1).All(&imageGoods)
		value["textgoods"]=textGoods
		value["imagegoods"]=imageGoods

	}

	this.Data["goods"]=goods
	tools.ShowLayout(&this.Controller)
	cartCount := tools.GetCartCount(&this.Controller)
	this.Data["cartCount"]=cartCount
	this.TplName="index.html"
}

//展示商品详情页
func(this *ShopController)ShowGoodsDetails(){
	id, err1 := this.GetInt("id")
	if err1!=nil{
	beego.Error("浏览器请求错误")
	this.Redirect("/index",302)
		return
	}
	o := orm.NewOrm()
	var goodsSKU models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType","Goods").Filter("Id",id).One(&goodsSKU)
	//获取时间靠前的同类型的2中商品
	var goodsNewSKU []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType",goodsSKU.GoodsType).OrderBy("Time").Limit(2,0).All(&goodsNewSKU)
	this.Data["goodsNewSKU"]=goodsNewSKU
	this.Data["goodsSKU"]=goodsSKU

	//添加历史浏览记录
	userName := this.GetSession("userName")
	if userName!=nil{
		//查询用户信息
		o := orm.NewOrm()
		user := models.User{}
		user.Name=userName.(string)
		o.Read(&user,"Name")
		//添加记录用redis存储
		conn, err1 := redis.Dial("tcp", "127.0.0.1:6379")
		if err1!=nil{
			beego.Info("redis链接错误")
		}
		defer conn.Close()
		//把以前相同的历史浏览记录删除
		conn.Do("lrem","history_"+strconv.Itoa(user.Id),0,id)
		//插入浏览的记录Id值
		conn.Do("lpush","history_"+strconv.Itoa(user.Id),id)
	}


	//判断用户是否登录
	tools.ShowLayout(&this.Controller)
	cartCount := tools.GetCartCount(&this.Controller)
	this.Data["cartCount"]=cartCount
	this.TplName="detail.html"
}


//展示菜单页面
func (this *ShopController)ShowList(){
	//获取数据
	id, err1 := this.GetInt("typeId")
	if err1!=nil{
		beego.Info("获取typeId失败")
		this.Redirect("/index",302)
		return
	}
	//获取type类型数据与传递视图
	tools.ShowLayout(&this.Controller)
	//获取最新商品2个
	o := orm.NewOrm()
	goodsNew := []models.GoodsSKU{}
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Time").Limit(2,0).All(&goodsNew)

	//获取展示同类商品
	goodsshow := []models.GoodsSKU{}
	//o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).All(&goodsshow)

	//分页功能实现
			//1)设置每页显示一条数据
			pagesize:=3
			//2)查询获取总共有多少条数据
			count, err2 := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).Count()
			if err2!=nil{
				beego.Info("获取总记录数失败")
			}
			//3)设置总页码数=总数据条数count/每页显示的条数pagesize	 转换float64格式 并使用天花板函数math.Ceil取整
			pageCount:=math.Ceil(float64(count)/float64(pagesize))
			//4)获取前端页面传过来的页码数
			pageindex, err3 := this.GetInt("pageindex")
			if err3!=nil{
				pageindex=1
			}
			//5)调用tools包中的分页函数, 得到pages数组
			pages := tools.PageTool(int(pageCount), pageindex)
			//6)传递数据给视图
			this.Data["pages"]=pages
			//7)将id传回给函数,解决点击页码跳转到首页的问题
			this.Data["typeId"]=id
			//8)把index值再传回给视图,解决页码被选中的问题 ,视图上做判断
			this.Data["pageindex"]=pageindex
			//9)设置显示的数据条数
			start:=(pageindex-1)*pagesize

			//获取上一页
			prepage:=pageindex-1
			if prepage<=1 {
				prepage=1
			}
			this.Data["prepage"]=prepage
			//获取下一页
			nextpage:=pageindex+1
			if nextpage >int(pageCount){
				nextpage=int(pageCount)
			}
			this.Data["nextpage"]=nextpage

	//商品热度排序
		sort:=this.GetString("sort")
		if sort==""{
			o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).Limit(pagesize,start).All(&goodsshow)
			this.Data["sort"]=""
			this.Data["goodsshow"]=goodsshow
		}else if sort=="price"{
			o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Price").Limit(pagesize,start).All(&goodsshow)
			this.Data["sort"]="price"
			this.Data["goodsshow"]=goodsshow

		}else {
			o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Sales").Limit(pagesize,start).All(&goodsshow)
			this.Data["sort"]="sale"
			this.Data["goodsshow"]=goodsshow

		}




	//传递视图

	this.Data["goodsNew"]=goodsNew

	this.TplName="list.html"
}

//搜素功能
func(this*ShopController)HandleSeach(){
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	goodsName := this.GetString("goodsName")
	beego.Info(goodsName)
	if goodsName==""{
		o.QueryTable("GoodsSKU").All(&goods)
		this.Data["goods"]=goods
		tools.ShowLayout(&this.Controller)
		this.TplName="seach.html"
		return
	}
	//处理数据
	o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).All(&goods)
	this.Data["goods"]=goods
	tools.ShowLayout(&this.Controller)
	this.TplName="seach.html"

}

