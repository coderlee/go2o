/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2013-12-12 16:55
 * description :
 * history :
 */

package merchant

import (
	"fmt"
	"go2o/core/domain/interface/merchant"
	"go2o/core/domain/interface/merchant/mss"
	"go2o/core/domain/interface/merchant/shop"
	"go2o/core/domain/interface/merchant/user"
	mssImpl "go2o/core/domain/merchant/mss"
	si "go2o/core/domain/merchant/shop"
	userImpl "go2o/core/domain/merchant/user"
	"go2o/core/infrastructure"
	"go2o/core/infrastructure/domain"
	"go2o/core/variable"
	"time"
)

var _ merchant.IMerchant = new(MerchantImpl)

type MerchantImpl struct {
	_value           *merchant.Merchant
	_saleConf        *merchant.SaleConf
	_host            string
	_rep             merchant.IMerchantRep
	_shopRep         shop.IShopRep
	_userRep         user.IUserRep
	_userManager     user.IUserManager
	_confManager     merchant.IConfManager
	_levelManager    merchant.ILevelManager
	_kvManager       merchant.IKvManager
	_memberKvManager merchant.IKvManager
	_mssManager      mss.IMssManager
	_mssRep          mss.IMssRep
	_profileManager  merchant.IProfileManager
	_apiManager      merchant.IApiManager
	_shopManager     shop.IShopManager
}

func NewMerchant(v *merchant.Merchant, rep merchant.IMerchantRep,
	shopRep shop.IShopRep, userRep user.IUserRep,
	mssRep mss.IMssRep) (merchant.IMerchant, error) {
	mch := &MerchantImpl{
		_value:   v,
		_rep:     rep,
		_shopRep: shopRep,
		_userRep: userRep,
		_mssRep:  mssRep,
	}
	return mch, mch.Stat()
}

func (this *MerchantImpl) GetRep() merchant.IMerchantRep {
	return this._rep
}

func (this *MerchantImpl) GetAggregateRootId() int {
	return this._value.Id
}
func (this *MerchantImpl) GetValue() merchant.Merchant {
	return *this._value
}

func (this *MerchantImpl) SetValue(v *merchant.Merchant) error {
	tv := this._value
	if v.Id == tv.Id {
		tv.Name = v.Name
		tv.Province = v.Province
		tv.City = v.City
		tv.District = v.District
		if v.LastLoginTime > 0 {
			tv.LastLoginTime = v.LastLoginTime
		}
		if v.LoginTime > 0 {
			tv.LoginTime = v.LoginTime
		}

		if len(v.Logo) != 0 {
			tv.Logo = v.Logo
		}
		tv.Pwd = v.Pwd
		tv.UpdateTime = time.Now().Unix()
	}
	return nil
}

// 保存
func (this *MerchantImpl) Save() (int, error) {
	var id int = this.GetAggregateRootId()
	if id > 0 {
		return this._rep.SaveMerchant(this._value)
	}

	return this.createMerchant()
}

// 获取商户的状态,判断 过期时间、判断是否停用。
// 过期时间通常按: 试合作期,即1个月, 后面每年延长一次。或者会员付费开通。
func (this *MerchantImpl) Stat() error {
	if this._value == nil {
		return merchant.ErrNoSuchMerchant
	}
	if this._value.Enabled == 0 {
		return merchant.ErrMerchantDisabled
	}
	if this._value.ExpiresTime < time.Now().Unix() {
		return merchant.ErrMerchantExpires
	}
	return nil
}

// 返回对应的会员编号
func (this *MerchantImpl) Member() int {
	return this.GetValue().MemberId
}

// 创建商户
func (this *MerchantImpl) createMerchant() (int, error) {
	if id := this.GetAggregateRootId(); id > 0 {
		return id, nil
	}

	v := this._value
	id, err := this._rep.SaveMerchant(v)
	if err != nil {
		return id, err
	}

	//todo:事务

	// 初始化商户信息
	this._value.Id = id

	//todo:  初始化商店

	// SiteConf
	//this._siteConf = &shop.ShopSiteConf{
	//	IndexTitle: "线上商店-" + v.Name,
	//	SubTitle:   "线上商店-" + v.Name,
	//	Logo:       v.Logo,
	//	State:      1,
	//	StateHtml:  "",
	//}
	//err = this._rep.SaveSiteConf(id, this._siteConf)
	//this._siteConf.MerchantId = id

	// SaleConf
	this._saleConf = &merchant.SaleConf{
		AutoSetupOrder:  1,
		IntegralBackNum: 0,
	}
	err = this._rep.SaveSaleConf(id, this._saleConf)
	this._saleConf.MerchantId = id

	// 创建API
	api := &merchant.ApiInfo{
		ApiId:     domain.NewApiId(id),
		ApiSecret: domain.NewSecret(id),
		WhiteList: "*",
		Enabled:   1,
	}
	err = this.ApiManager().SaveApiInfo(api)
	return id, err
}

// 获取商户的域名
func (this *MerchantImpl) GetMajorHost() string {
	if len(this._host) == 0 {
		host := this._rep.GetMerchantMajorHost(this.GetAggregateRootId())
		if len(host) == 0 {
			host = fmt.Sprintf("%s.%s", this._value.Usr, infrastructure.GetApp().
				Config().GetString(variable.ServerDomain))
		}
		this._host = host
	}
	return this._host
}

// 获取销售配置
func (this *MerchantImpl) GetSaleConf() merchant.SaleConf {
	if this._saleConf == nil {
		//10%分成
		//0.2,         #上级
		//0.1,         #上上级
		//0.8          #消费者自己
		this._saleConf = this._rep.GetSaleConf(
			this.GetAggregateRootId())

		this.verifySaleConf(this._saleConf)
	}
	return *this._saleConf
}

// 保存销售配置
func (this *MerchantImpl) SaveSaleConf(v *merchant.SaleConf) error {
	this.GetSaleConf()
	if v.FlowConvertCsn < 0 || v.PresentConvertCsn < 0 ||
		v.ApplyCsn < 0 || v.TransCsn < 0 ||
		v.FlowConvertCsn > 1 || v.PresentConvertCsn > 1 ||
		v.ApplyCsn > 1 || v.TransCsn > 1 {
		return merchant.ErrSalesPercent
	}

	this.verifySaleConf(v)

	this._saleConf = v
	this._saleConf.MerchantId = this._value.Id

	return this._rep.SaveSaleConf(this.GetAggregateRootId(), this._saleConf)
}

// 验证销售设置
func (this *MerchantImpl) verifySaleConf(v *merchant.SaleConf) {
	if v.OrderTimeOutMinute <= 0 {
		v.OrderTimeOutMinute = 1440 // 一天
	}

	if v.OrderConfirmAfterMinute <= 0 {
		v.OrderConfirmAfterMinute = 60 // 一小时后自动确认
	}

	if v.OrderTimeOutReceiveHour <= 0 {
		v.OrderTimeOutReceiveHour = 7 * 24 // 7天后自动确认
	}
}

// 返回用户服务
func (this *MerchantImpl) UserManager() user.IUserManager {
	if this._userManager == nil {
		this._userManager = userImpl.NewUserManager(
			this.GetAggregateRootId(),
			this._userRep)
	}
	return this._userManager
}

// 返回设置服务
func (this *MerchantImpl) ConfManager() merchant.IConfManager {
	if this._confManager == nil {
		this._confManager = &ConfManager{
			_merchantId: this.GetAggregateRootId(),
			_rep:        this._rep,
		}
	}
	return this._confManager
}

// 获取会员管理服务
func (this *MerchantImpl) LevelManager() merchant.ILevelManager {
	return nil
	/*
	   if this._levelManager == nil {
	       this._levelManager = NewLevelManager(this.GetAggregateRootId(), this._memberRep)
	   }
	   return this._levelManager*/
}

// 获取键值管理器
func (this *MerchantImpl) KvManager() merchant.IKvManager {
	if this._kvManager == nil {
		this._kvManager = newKvManager(this, "kvset")
	}
	return this._kvManager
}

// 获取用户键值管理器
func (this *MerchantImpl) MemberKvManager() merchant.IKvManager {
	if this._memberKvManager == nil {
		this._memberKvManager = newKvManager(this, "kvset_member")
	}
	return this._memberKvManager
}

// 消息系统管理器
func (this *MerchantImpl) MssManager() mss.IMssManager {
	if this._mssManager == nil {
		this._mssManager = mssImpl.NewMssManager(this, this._mssRep, this._rep)
	}
	return this._mssManager
}

// 企业资料管理器
func (this *MerchantImpl) ProfileManager() merchant.IProfileManager {
	if this._profileManager == nil {
		this._profileManager = newProfileManager(this)
	}
	return this._profileManager
}

// API服务
func (this *MerchantImpl) ApiManager() merchant.IApiManager {
	if this._apiManager == nil {
		this._apiManager = newApiManagerImpl(this)
	}
	return this._apiManager
}

// 商店服务
func (this *MerchantImpl) ShopManager() shop.IShopManager {
	if this._shopManager == nil {
		this._shopManager = si.NewShopManagerImpl(this, this._shopRep)
	}
	return this._shopManager
}
