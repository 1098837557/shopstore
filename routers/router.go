package routers

import (
    "github.com/astaxie/beego/context"
    "shopstore/controllers"
	"github.com/astaxie/beego"
)

func init() {
    //路由器过滤函数

    beego.InsertFilter("/store/*",beego.BeforeRouter,FilterFunc)

    //注册路由
    beego.Router("/register",&controllers.UserController{},"get:ShowReg;post:HandleReg")
    //激活路由
    beego.Router("/active",&controllers.UserController{},"get:ActiveUser")
    //登录路由
    beego.Router("/login",&controllers.UserController{},"get:ShowLog;post:HandleLog")
    //访问菜单路由
    beego.Router("/list",&controllers.ShopController{},"get:ShowList")
    //展示首页
    beego.Router("/index",&controllers.ShopController{},"get:ShowIndex")
    //访问详情页路由
    beego.Router("/goodsDetails",&controllers.ShopController{},"get:ShowGoodsDetails")




    //用户信息中心页面
    beego.Router("/store/usercenterinfo",&controllers.UserController{},"get:ShowUserCenterInfo")
    //订单中心页面
    beego.Router("/store/usercenterorder",&controllers.UserController{},"get:ShowUserCenterOrder")

    //收货地址页面
    beego.Router("/store/usercentersite",&controllers.UserController{},"get:ShowUserCenterSite;post:HandleUserCenterSite")
    //添加购物车路由
    beego.Router("/store/addCart",&controllers.CartController{},"post:HandleAddCart")

    //显示处理购物车路由器
    beego.Router("/store/cart",&controllers.CartController{},"get:ShowCart")

    //商品搜索路由
    beego.Router("/goodsseach",&controllers.ShopController{},"post:HandleSeach")

    //退出登录路由
    beego.Router("/store/quitlogin",&controllers.UserController{},"get:QuitLogin")

}
var FilterFunc= func(ctx *context.Context) {
    username:=ctx.Input.Session("userName")
    if username==nil{
        ctx.Redirect(302,"/login")
    }
}