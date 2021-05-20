package dao

import (
	"bilibili/module"
	"bilibili/tool"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
)

type Userdao struct {

}

func (ud *Userdao)CreateTableUser(user *module.User){
	err:=tool.DBEngine.AutoMigrate(user).Error
	if err!=nil{
		panic(err)
	}
}

//根据用户名查找并返回整个结构体
func (ud *Userdao)FindUser(userpoint *module.User)(*module.User,error){
	oldUser:=new(module.User)
	err:=tool.DBEngine.Where(userpoint).First(oldUser).Error

	fmt.Printf("%#v\n",*userpoint)
	 if err!=nil{
		fmt.Println("find err:",err)
		return nil,err  //err:record not find
	}
	return oldUser,nil
}


//传入结构体添加入用户数据表
func (ud *Userdao)CreateUser(user *module.User)error{
	err:=tool.DBEngine.Create(user).Error
	if err!=nil{
		fmt.Println("create user err:",err)
		return err
	}
	return nil
}

func (ud *Userdao)ChangeUser(olduser,newuser *module.User)error{
	if olduser==nil{
		err:=errors.New("olduser is nil")
		return err
	}

	err:=tool.DBEngine.Model(&module.User{}).Where(olduser).Updates(newuser).Error
	if err!=nil{
		fmt.Println("updata err:",err)
		return err
	}
	return nil
}

//开启事务
func (ud *Userdao)ChangeUserByAffairs(tx *gorm.DB,olduser,newuser *module.User)error{
	if olduser==nil{
		err:=errors.New("olduser is nil")
		return err
	}

	err:=tx.Model(&module.User{}).Where(olduser).Updates(newuser).Error
	if err!=nil{
		fmt.Println("updata err:",err)
		return err
	}
	return nil
}


func (ud *Userdao)CreateAnimals(olduser,newuser *module.User)error{
	var typeInfo = reflect.TypeOf(*newuser)
	var valInfo = reflect.ValueOf(*newuser)
	num := typeInfo.NumField()
	fmt.Println(typeInfo,valInfo,num,olduser.Uid)//test

	var index []int
	for i := 0; i < num; i++ {
		//找到需要查询的字段
		val:= valInfo.Field(i).Interface()
		if val!=nil&&val!=0&&val!=""{
			index=append(index,i)
		}
	}
	if len(index)==0{
		err:=errors.New("field index out of bounds")
		return err
	}

	tx:=tool.DBEngine.Begin()//开启事务
	//一旦在一个事务中，需使用tx作为数据库句柄
	err:=errors.New("")
	for _,v:=range index {
		if valInfo.Field(v).Interface()==-1 {
			err=tx.Model(olduser).Where("uid=?", olduser.Uid).Omit("uid").Update(typeInfo.Field(v).Name, 0).Error
		}else {
			err= tx.Model(olduser).Where("uid=?", olduser.Uid).Omit("uid").Update(typeInfo.Field(v).Name, valInfo.Field(v).Interface()).Error
		}

		if err!=nil{
			tx.Rollback()
			fmt.Println("tx can not commit")
			return err
		}
	}
	tx.Commit()//事务提交
	return nil
}


//用户关注
func (ud Userdao)InsertFollows(uid ,folloid,username string)error{
	Follows:=module.Follows{
		Uid: uid,
		FollowId: folloid,
		FollowName: username,
	}
	err:=tool.DBEngine.Create(&Follows).Error
	if err!=nil{
		fmt.Println("create follows err:",err)
		return err
	}
	return nil
}

func (ud *Userdao)DeleteFollows(uid,followid string)error{
	return tool.DBEngine.Where(&module.Follows{Uid: uid,FollowId: followid}).Delete(&module.Follows{}).Error
}

func (ud *Userdao)FindFollowID(uid,followid string)error{
	var follows module.Follows
	return tool.DBEngine.Where(&module.Follows{Uid: uid,FollowId: followid}).First(&follows).Error
}


//用户与视频关联
func (ud Userdao)FindLikes(newLikes *module.UserWithVideo) (*[]module.UserWithVideo, error) {
	oldLikes:=new([]module.UserWithVideo)
	err:=tool.DBEngine.Model(&module.UserWithVideo{}).Where(newLikes).Find(oldLikes).Error
	return oldLikes,err
}

func (ud Userdao)CreateLikes(newLikes *module.UserWithVideo)error{
	return tool.DBEngine.Create(newLikes).Error
}

func (ud Userdao)DeleteLikes(newLikes *module.UserWithVideo)error{
	return tool.DBEngine.Model(&module.UserWithVideo{}).Where(newLikes).Delete(newLikes).Error
}

func (ud Userdao)UpdateLikes(oldLikes,newLikes *module.UserWithVideo)error{
	return tool.DBEngine.Where(oldLikes).Update(newLikes).Error
}

