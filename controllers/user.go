package controllers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/garyburd/redigo/redis"
	"regexp"
	"shopstore/models"
	"strconv"
	"shopstore/tools"
)

type UserController struct {
	beego.Controller
}
//展示注册页面
func(this *UserController)ShowReg(){

	this.TplName="register.html"
}
//处理注册请求
func (this *UserController)HandleReg(){
	username := this.GetString("user_name")
	password := this.GetString("pwd")
	cpassword := this.GetString("cpwd")
	email := this.GetString("email")
	if username==""||password==""||cpassword==""||email==""{
		this.Data["errmsg"]="注册的信息有误,请重新输入"
		this.TplName="register.html"
		return
	}
	if password!=cpassword{
		this.Data["errmsg"]="两次密码不一致,请重新输入"
		this.TplName="register.html"
		return
	}
	//邮箱格式判断,用正则表达式
	reg, _ := regexp.Compile("^[a-z0-9A-Z]+[- | a-z0-9A-Z . _]+@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-z]{2,}$")
	res := reg.FindString(email)
	if res==""{
		this.Data["errmsg"]="邮箱格式不正确,请重新输入"
		this.TplName="register.html"
		return
	}

	//插入数据
	user := models.User{}
	o := orm.NewOrm()
	user.Name=username
	user.PassWord=password
	user.Email=email
	_, err1 := o.Insert(&user)
	if err1!=nil{
		this.Data["errmsg"]="注册失败,请重新注册"
		this.TplName="register.html"
		return
	}
	//发送邮件激活
	emailConfig:=`{"username":"1098837557@qq.com","password":"ugrocwimbnpnhfgd","host":"smtp.qq.com","port":587}`
	emailConn := utils.NewEMail(emailConfig)
	emailConn.From="天天生鲜注册服务"
	//一次可以发好多人,所以用切片
	emailConn.To=[]string{email}
	emailConn.Subject="天天生鲜用户注册"
	//这里给用户发送的激活请求地址
	emailConn.Text="172.20.10.6:9090/active/?id="+strconv.Itoa(user.Id)
	err2 := emailConn.Send()
	if err2!=nil{
		beego.Info(err2,"发送邮件失败了")
	}

	//返回试图
	this.Ctx.WriteString("注册成功!,请去邮箱激活该账号.")


}
//激活用户处理
func(this *UserController)ActiveUser(){
	//获取数据
	id, err1 := this.GetInt("id")
	//校验数据
	if err1!=nil{
		this.Data["errmsg"]="要激活的用户不存在"
		this.TplName="register.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	user := models.User{}
	user.Id=id
	err2 := o.Read(&user, "Id")
	if err2!=nil{
		this.Data["errmsg"]="要激活的用户不存在"
		this.TplName="register.html"
		return
	}
	user.Active=true
	o.Update(&user)
	//返回试图
	this.Redirect("/login",302)


}
//展示登录页面
func(this *UserController)ShowLog(){
	//展示页面同时获取cooke数据
	username := this.Ctx.GetCookie("username")
	//解密加密的数据
	tempusername, err := base64.StdEncoding.DecodeString(username)
	if err !=nil{
		beego.Info("数据解密失败")
	}
	if string(tempusername)==""{
		this.Data["userName"]=""
		this.Data["checked"]=""
	}else {
		this.Data["userName"]=string(tempusername)
		this.Data["checked"]="checked"
	}
	this.TplName="login.html"
}
//登录处理
func (this *UserController)HandleLog(){
	//获取数据
	username := this.GetString("username")
	pwd := this.GetString("pwd")
	//判断数据
  if username==""||pwd==""{
  	this.Data["errmsg"]="用户名或密码不能为空!"
  	this.TplName="login.html"
	  return
  }
	//对比数据
	o := orm.NewOrm()
	user := models.User{}
	user.Name=username
	err1 := o.Read(&user, "Name")
	if err1!=nil{
		this.Data["errmsg"]="用户名或密码不正确,请重新输入"
		beego.Info("用户不存在")
		this.TplName="login.html"
		return
	}
	if user.PassWord!=pwd{
		this.Data["errmsg"]="用户名或密码不正确,请重新输入"
		beego.Info("密码错误")
		this.TplName="login.html"
		return
	}
	//设置用户访问页面判断seesion
	this.SetSession("userName",username)

	//设置记住用户名
	remember := this.GetString("remember")
	if remember=="on"{
		//设置cookie之前先数据加密
		tempUserName := base64.StdEncoding.EncodeToString([]byte(username))
		//设置cookie
		this.Ctx.SetCookie("username",tempUserName,3600*24)

	}else {
		this.Ctx.SetCookie("username",username,-1)
	}

	//跳转视图
	this.Redirect("/index",302)
}

//显示用户信息中心页面
func (this *UserController) ShowUserCenterInfo()  {
	//导入tools获取用户名的函数
	tools.GetUser(&this.Controller)
	username := this.GetSession("userName")
	this.Data["username"]=username.(string)
	//查询地址表 表关联查询
	o := orm.NewOrm()
	var addr models.Address
	//以user表中的user-name为查询条件.查询一整条完整的address数据,存放到addr中
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username.(string)).Filter("Isdefault", true).One(&addr)
	if addr.Id==0{
		this.Data["addr"]=""
	}else {
		this.Data["addr"]=addr
	}
	//获取浏览历史记录
	conn, err2 := redis.Dial("tcp", "127.0.0.1:6379")
	if err2!=nil{
		beego.Info("链接redis错误")
	}
	user := models.User{}
	var goodsSKUs []models.GoodsSKU
	user.Name=username.(string)
	o.Read(&user,"Name")
	//返回的是一个接口类型数据
	reply, err2 := conn.Do("lrange", "history_"+strconv.Itoa(user.Id), 0, 4)
	defer conn.Close()
	//回复助手函数  把接口转换成相应的数据类型 //得到goodsIds数组
	goodsIds, _ := redis.Ints(reply, err2)
	//遍历goodsIds数组
	for _,values :=range goodsIds{
		//创建goodsSKU实例
		var goodsSKU models.GoodsSKU
		//Id赋值
		goodsSKU.Id=values
		//读取数据
		o.Read(&goodsSKU,"Id")
		//append方法向数组中添加元素
		goodsSKUs=append(goodsSKUs,goodsSKU)
	}


	this.Data["goodsSKUs"]=goodsSKUs
	this.Layout="layout.html"
	this.TplName="user_center_info.html"
}

//显示用户订单页面
func(this *UserController)ShowUserCenterOrder(){
	tools.GetUser(&this.Controller)
	this.Layout="layout.html"
	this.TplName="user_center_order.html"
}

//显示收货地址页面
func (this *UserController)ShowUserCenterSite(){
	tools.GetUser(&this.Controller)
	username := this.GetSession("userName")
	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username.(string)).Filter("Isdefault",true).One(&addr)
	this.Data["addr"]=addr
	this.Layout="layout.html"
	this.TplName="user_center_site.html"
}
//更新地址处理
func (this *UserController)HandleUserCenterSite(){
	receiver := this.GetString("Receiver")
	addr := this.GetString("addr")
	zip_code := this.GetString("zip_code")
	phone:= this.GetString("phone")
	if receiver==""||addr==""||zip_code==""||phone==""{
		this.Redirect("/store/usercentersite",302)
		return
	}
	o := orm.NewOrm()
	address := models.Address{}
	address.Isdefault=true
	err1 := o.Read(&address, "Isdefault")
	if err1 ==nil{
		address.Isdefault=false
		o.Update(&address)
	}

	//关联user表
	username := this.GetSession("userName")
	var user models.User
	user.Name=username.(string)
	o.Read(&user,"Name")
	//插入数据
	var addrNew models.Address
	addrNew.Receiver=receiver
	addrNew.Addr=addr
	addrNew.Zipcode=zip_code
	addrNew.Phone=phone
	addrNew.Isdefault=true
	addrNew.User=&user
	o.Insert(&addrNew)

	//跳转视图
	this.Redirect("/store/usercentersite",302)


}

//退出登录
func (this *UserController)QuitLogin(){
	this.DelSession("userName")
	this.Redirect("/login",302)
}