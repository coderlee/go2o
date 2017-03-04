/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2013-12-05 17:37
 * description :
 * history :
 */

package order

import (
	"go2o/core/domain/interface/cart"
	"go2o/core/domain/interface/member"
	"go2o/core/domain/interface/payment"
	"go2o/core/domain/interface/promotion"
	"go2o/core/infrastructure/domain"
)

// 自动拆单应在下单前完成
// 用户拆单,则重新生成子订单
// 参考:
//http://www.pmcaff.com/discuss?id=1000000000138488
//http://www.zhihu.com/question/31640837

// 拆分为子订单单暂时不拆分支付单

type OrderState int
type OrderType int32

const (
	// 零售订单(线上/线下)
	TRetail OrderType = 1
	// 批发订单
	TWholesale OrderType = 2
	// 虚拟订单,如：手机充值
	TVirtual OrderType = 3
	// 交易订单,如：线下订单支付。
	TTrade OrderType = 4
	// 服务订单
	TService OrderType = 5
)

const (
	/****** 在履行前,订单可以取消申请推狂  ******/

	// 等待支付
	StatAwaitingPayment = 1
	// 等待确认
	StatAwaitingConfirm = 2
	// 等待备货
	StatAwaitingPickup = 3
	// 等待发货
	StatAwaitingShipment = 4

	/****** 订单取消 ******/

	// 系统取消
	StatCancelled = 11
	// 买家申请取消,等待卖家确认
	StatAwaitingCancel = 12
	// 卖家谢绝订单,由于无货等原因
	StatDeclined = 13
	// 已退款,完成取消
	StatRefunded = 14

	/****** 履行后订单只能退货或换货 ******/

	// 部分发货(将订单商品分多个包裹发货)
	PartiallyShipped = 5
	// 完成发货
	StatShipped = 6
	// 订单完成
	StatCompleted = 7
	// 订单已拆分
	StatBreak = 8

	/****** 售后状态 ******/

	// 已退货
	StatGoodsRefunded = 15
)

func (t OrderState) String() string {
	switch t {
	case StatAwaitingPayment:
		return "待付款"
	case StatAwaitingConfirm:
		return "待确认"
	case StatAwaitingPickup:
		return "正在备货"
	case StatAwaitingShipment:
		return "等待发货"
	case StatCancelled:
		return "交易关闭"
	case StatDeclined:
		return "卖家关闭"
	case StatAwaitingCancel:
		return "等待退款"
	case StatRefunded:
		return "已退款"
	case PartiallyShipped:
		return "已部分发货"
	case StatShipped:
		return "待收货"
	case StatCompleted:
		return "交易完成"
	case StatGoodsRefunded:
		return "已退货"
	}
	return "Error State"
}

// 后端状态描述
func (t OrderState) BackEndString() string {
	return t.String()
}

const (
	LogSetup       LogType = 1
	LogChangePrice LogType = 2
)

type LogType int

func (o LogType) String() string {
	switch o {
	case LogSetup:
		return "流程"
	case LogChangePrice:
		return "调价"
	}
	return ""
}

var (
	ErrRequireCart *domain.DomainError = domain.NewDomainError(
		"err_require_cart", "订单已生成,无法引入购物车")

	ErrNoSuchOrder *domain.DomainError = domain.NewDomainError(
		"err_no_such_order", "订单不存在")

	ErrOrderPayed *domain.DomainError = domain.NewDomainError(
		"err_order_payed ", "订单已支付")

	ErrNoYetCreated *domain.DomainError = domain.NewDomainError(
		"err_order_not_yet_created ", "订单尚未生成")

	ErrUnusualOrder *domain.DomainError = domain.NewDomainError(
		"err_unusual_order", "订单异常")

	ErrMissingShipAddress *domain.DomainError = domain.NewDomainError(
		"err_missing_ship_address", "未设置收货地址")

	ErrUnusualOrderStat *domain.DomainError = domain.NewDomainError(
		"err_except_order_stat", "订单状态不匹配、无法执行此操作!")

	ErrPartialShipment *domain.DomainError = domain.NewDomainError(
		"err_order_partial_shipment", "订单部分商品已经发货")

	ErrOrderNotPayed *domain.DomainError = domain.NewDomainError(
		"err_order_not_payed ", "订单未支付")

	ErrOutOfQuantity *domain.DomainError = domain.NewDomainError(
		"err_order_out_of_quantity", "超出数量")
	ErrNoSuchGoodsOfOrder *domain.DomainError = domain.NewDomainError(
		"err_order_no_such_goods_of_order", "订单中不包括该商品")
	ErrOrderHasConfirm *domain.DomainError = domain.NewDomainError(
		"err_order_has_confirm", "订单已经确认")

	ErrOrderNotConfirm *domain.DomainError = domain.NewDomainError(
		"err_order_not_confirm", "请等待系统确认")

	ErrOrderHasPickUp *domain.DomainError = domain.NewDomainError(
		"err_order_has_pick_up", "订单已经备货")

	ErrOrderNotPickUp *domain.DomainError = domain.NewDomainError(
		"err_order_not_pick_up", "请等待商品备货")

	ErrNoSuchAddress *domain.DomainError = domain.NewDomainError(
		"err_order_no_address", "请选择收货地址")

	ErrOrderShipped *domain.DomainError = domain.NewDomainError(
		"err_order_shipped", "订单已经发货")

	ErrOrderNotShipped *domain.DomainError = domain.NewDomainError(
		"err_order_not_shipped", "订单尚未发货")

	ErrIsCompleted *domain.DomainError = domain.NewDomainError(
		"err_order_is_completed", "订单已经完成")

	ErrOrderBreakUpFail *domain.DomainError = domain.NewDomainError(
		"err_order_break_up_fail", "拆分订单操作失败")

	ErrPromotionApplied *domain.DomainError = domain.NewDomainError(
		"err_promotion_applied", "已经使用相同的促销")

	ErrEmptyReason *domain.DomainError = domain.NewDomainError(
		"err_order_empty_reason", "原因不能为空")

	ErrOrderCancelled *domain.DomainError = domain.NewDomainError(
		"err_order_can_not_cancel", "订单已经取消")

	ErrOrderShippedCancel *domain.DomainError = domain.NewDomainError(
		"err_order_shipped_cancel", "订单已发货，无法取消")

	ErrHasRefund *domain.DomainError = domain.NewDomainError(
		"err_order_has_refund", "订单已经退款")

	ErrDisallowRefund *domain.DomainError = domain.NewDomainError(
		"err_order_disallow_refund", "订单不允许退款")
)

type (
	IOrder interface {
		// 获取编号
		GetAggregateRootId() int64
		// 订单类型
		Type() OrderType
		// 获取订单状态
		State() OrderState
		// 获取购买的会员
		Buyer() member.IMember
		// 获取订单号
		OrderNo() string
		// 复合的订单信息
		Complex() *ComplexOrder
		// 提交订单。如遇拆单,需均摊优惠抵扣金额到商品
		Submit() error
	}

	// 普通订单
	INormalOrder interface {
		// 读取购物车数据,用于预生成订单
		RequireCart(c cart.ICart) error
		// 根据运营商获取商品和运费信息,限未生成的订单
		GetByVendor() (items map[int32][]*SubOrderItem, expressFee map[int32]float32)
		// 获取支付单
		GetPaymentOrder() payment.IPaymentOrder
		// 在线支付交易完成
		OnlinePaymentTradeFinish() error
		// 设置配送地址
		SetAddress(addressId int32) error
		// 提交订单。如遇拆单,需均摊优惠抵扣金额到商品
		Submit() error

		//根据运营商拆单,返回拆单结果,及拆分的订单数组
		//BreakUpByVendor() ([]IOrder, error)
		// 获取子订单列表
		GetSubOrders() []ISubOrder
		// 应用优惠券
		ApplyCoupon(coupon promotion.ICouponPromotion) error
		// 获取应用的优惠券
		GetCoupons() []promotion.ICouponPromotion
		// 获取可用的促销,不包含优惠券
		GetAvailableOrderPromotions() []promotion.IPromotion
		// 获取最省的促销
		GetBestSavePromotion() (p promotion.IPromotion,
			saveFee float32, integral int)
		// 获取促销绑定
		GetPromotionBinds() []*OrderPromotionBind
	}

	// 子订单(普通订单拆分)
	ISubOrder interface {
		// 获取领域对象编号
		GetDomainId() int64
		// 获取值对象
		GetValue() *NormalSubOrder
		// 复合的订单信息
		Complex() *ComplexOrder

		// 获取商品项
		Items() []*SubOrderItem
		// 在线支付交易完成
		PaymentFinishByOnlineTrade() error
		// 记录订单日志
		AppendLog(logType LogType, system bool, message string) error
		// 添加备注
		AddRemark(string)
		// 确认订单
		Confirm() error
		// 捡货(备货)
		PickUp() error
		// 发货
		Ship(spId int32, spOrder string) error
		// 已收货
		BuyerReceived() error
		// 获取订单的日志
		LogBytes() []byte
		// 挂起
		Suspend(reason string) error
		// 取消订单/退款
		Cancel(reason string) error
		// 退回商品
		Return(snapshotId int32, quantity int32) error
		// 撤销退回商品
		RevertReturn(snapshotId int32, quantity int32) error
		// 谢绝订单
		Decline(reason string) error
		// 提交子订单
		Submit() (int64, error)
	}

	// 批发订单
	IWholesaleOrder interface {
		// 获取领域对象编号
		//GetDomainId() int64
		// 获取值对象
		//GetValue() *NormalSubOrder

		// 设置商品项
		SetItems(items []*MinifyItem)
		// 设置配送地址
		SetAddress(addressId int32) error
		// 获取商品项
		Items() []*WholesaleItem
		// 获取支付单
		GetPaymentOrder() payment.IPaymentOrder
		// 在线支付交易完成
		OnlinePaymentTradeFinish() error

		// 记录订单日志
		AppendLog(logType LogType, system bool, message string) error
		// 添加备注
		AddRemark(string)
		// 确认订单
		Confirm() error
		// 捡货(备货)
		PickUp() error
		// 发货
		Ship(spId int32, spOrder string) error
		// 已收货
		BuyerReceived() error
		// 获取订单的日志
		LogBytes() []byte
		// 取消订单/退款
		Cancel(reason string) error
		// 谢绝订单
		Decline(reason string) error
		// 退回商品
		//Return(snapshotId int32, quantity int32) error
		// 撤销退回商品
		//RevertReturn(snapshotId int32, quantity int32) error
	}

	// 交易订单
	ITradeOrder interface{}

	// 订单
	Order struct {
		ID int64 `db:"id" pk:"yes" auto:"yes"`
		// 订单号
		OrderNo string `db:"order_no"`
		// 买家编号
		BuyerId int32 `db:"buyer_id"`
		// 订单类型
		OrderType int32 `db:"order_type"`
		// 订单状态
		State int32 `db:"state"`
		// 下单时间
		CreateTime int64 `db:"create_time"`
	}

	// 订单复合信息
	ComplexOrder struct {
		// 订单编号
		OrderId int64
		// 子订单编号
		SubOrderId int64
		// 订单类型
		OrderType int32
		// 订单号
		OrderNo string
		// 购买人编号
		BuyerId int32
		// 运营商编号
		VendorId int32
		// 店铺编号
		ShopId int32
		// 商品金额
		ItemAmount float64
		// 优惠减免金额
		DiscountAmount float64
		// 运费
		ExpressFee float64
		// 包装费用
		PackageFee float64
		// 实际金额
		FinalAmount float64
		// 收货人
		ConsigneePerson string
		// 收货人联系电话
		ConsigneePhone string
		// 收货地址
		ShippingAddress string
		// 订单是否拆分
		IsBreak int32
		// 订单状态
		State int32
		// 订单生成时间
		CreateTime int64
		// 更新时间
		UpdateTime int64
		// 商品项
		Items []*ComplexItem
	}
	// 符合的订单项
	ComplexItem struct {
		// 编号
		ID int64 `db:"id" pk:"yes" auto:"yes" json:"id"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 商品编号
		ItemId int64 `db:"item_id"`
		// 商品SKU编号
		SkuId int64 `db:"sku_id"`
		// 快照编号
		SnapshotId int64 `db:"snap_id"`
		// 数量
		Quantity int32 `db:"quantity"`
		// 退回数量(退货)
		ReturnQuantity int32 `db:"return_quantity"`
		// 金额
		Amount float64 `db:"amount"`
		// 最终金额, 可能会有优惠均摊抵扣的金额
		FinalAmount float64 `db:"final_amount"`
		// 是否发货
		IsShipped int32 `db:"is_shipped"`
	}

	// 订单商品项
	MinifyItem struct {
		ItemId   int32
		SkuId    int32
		Quantity int32
	}

	// 普通订单
	NormalOrder struct {
		// 编号
		ID int64 `db:"id" pk:"yes" auto:"yes"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 商品金额
		ItemAmount float32 `db:"item_amount"`
		// 优惠减免金额
		DiscountAmount float32 `db:"discount_amount" json:"discountFee"`
		// 运费
		ExpressFee float32 `db:"express_fee"`
		// 包装费用
		PackageFee float32 `db:"package_fee"`
		// 实际金额
		FinalAmount float32 `db:"final_amount" json:"fee"`
		// 收货人
		ConsigneePerson string `db:"consignee_person" json:"deliverName"`
		// 收货人联系电话
		ConsigneePhone string `db:"consignee_phone" json:"deliverPhone"`
		// 收货地址
		ShippingAddress string `db:"shipping_address" json:"deliverAddress"`
		// 订单是否拆分
		IsBreak int32 `db:"is_break"`
		// 更新时间
		UpdateTime int64 `db:"update_time" json:"updateTime"`
	}

	// 子订单
	NormalSubOrder struct {
		// 编号
		ID int64 `db:"id" pk:"yes" auto:"yes"`
		// 订单号
		OrderNo string `db:"order_no"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 购买人编号(冗余,便于商户处理数据)
		BuyerId int32 `db:"buyer_id"`
		// 运营商编号
		VendorId int32 `db:"vendor_id" json:"vendorId"`
		// 店铺编号
		ShopId int32 `db:"shop_id" json:"shopId"`
		// 订单标题
		Subject string `db:"subject" json:"subject"`
		// 商品金额
		ItemAmount float32 `db:"item_amount"`
		// 优惠减免金额
		DiscountAmount float32 `db:"discount_amount" json:"discountFee"`
		// 运费
		ExpressFee float32 `db:"express_fee"`
		// 包装费用
		PackageFee float32 `db:"package_fee"`
		// 实际金额
		FinalAmount float32 `db:"final_amount" json:"fee"`
		// 是否支付
		IsPaid int `db:"is_paid"`
		// 是否挂起，如遇到无法自动进行的时挂起，来提示人工确认。
		IsSuspend int `db:"is_suspend" json:"is_suspend"`
		// 顾客备注
		BuyerRemark string `db:"buyer_remark" json:"note"`
		// 系统备注
		Remark string `db:"remark" json:"remark"`
		// 订单状态
		State int32 `db:"state" json:"state"`
		// 下单时间
		CreateTime int64 `db:"create_time"`
		// 更新时间
		UpdateTime int64 `db:"update_time" json:"updateTime"`
		// 订单项
		Items []*SubOrderItem `db:"-"`
	}

	// 订单商品项
	SubOrderItem struct {
		// 编号
		ID int32 `db:"id" pk:"yes" auto:"yes" json:"id"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 商品编号
		ItemId int32 `db:"item_id"`
		// 商品SKU编号
		SkuId int32 `db:"sku_id"`
		// 快照编号
		SnapshotId int32 `db:"snap_id"`
		// 数量
		Quantity int32 `db:"quantity"`
		// 退回数量(退货)
		ReturnQuantity int32 `db:"return_quantity"`
		// SKU描述
		//Sku string `db:"sku"`
		// 金额
		Amount float32 `db:"amount"`
		// 最终金额, 可能会有优惠均摊抵扣的金额
		FinalAmount float32 `db:"final_amount"`
		// 是否发货
		IsShipped int `db:"is_shipped"`
		// 更新时间
		UpdateTime int64 `db:"update_time"`
		// 运营商编号
		VendorId int32 `db:"-"`
		// 商店编号
		ShopId int32 `db:"-"`
		// 重量,用于生成订单时存储数据
		Weight int32 `db:"-"`
		// 体积:毫升(ml)
		Bulk int32 `db:"-"`
		// 快递模板编号
		ExpressTplId int32 `db:"-"`
	}

	// 批发订单
	WholesaleOrder struct {
		// 编号
		ID int64 `db:"id" pk:"yes" auto:"yes"`
		// 订单号
		OrderNo string `db:"order_no"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 买家编号
		BuyerId int32 `db:"buyer_id"`
		// 商家编号
		VendorId int32 `db:"vendor_id"`
		// 店铺编号
		ShopId int32 `db:"shop_id"`
		// 商品总价
		ItemAmount float32 `db:"item_amount"`
		// 抵扣金额
		DiscountAmount float32 `db:"discount_amount"`
		// 运费
		ExpressFee float32 `db:"express_fee"`
		// 包装费
		PackageFee float32 `db:"package_fee"`
		// 订单最终金额
		FinalAmount float32 `db:"final_amount"`
		// 收货人姓名
		ConsigneePerson string `db:"consignee_person"`
		// 收货人电话
		ConsigneePhone string `db:"consignee_phone"`
		// 收货人地址
		ShippingAddress string `db:"shipping_address"`
		// 是否支付
		IsPaid int32 `db:"is_paid"`
		// 订单备注
		Remark string `db:"remark"`
		// 订单买家备注
		BuyerRemark string `db:"buyer_remark"`
		// 订单状态
		State int32 `db:"state"`
		// 订单创建时间
		CreateTime int64 `db:"create_time"`
		// 订单更新时间
		UpdateTime int64 `db:"update_time"`
	}

	// 批发订单商品
	WholesaleItem struct {
		// 编号
		ID int64 `db:"id" pk:"yes" auto:"yes"`
		// 订单编号
		OrderId int64 `db:"order_id"`
		// 商品编号
		ItemId int64 `db:"item_id"`
		// SKU编号
		SkuId int64 `db:"sku_id"`
		// 商品快照编号
		SnapshotId int64 `db:"snapshot_id"`
		// 销售数量
		Quantity int32 `db:"quantity"`
		// 退货数量
		ReturnQuantity int32 `db:"return_quantity"`
		// 商品总金额
		Amount float32 `db:"amount"`
		// 商品实际金额
		FinalAmount float32 `db:"final_amount"`
		// 是否已发货
		IsShipped int32 `db:"is_shipped"`
		// 更新时间
		UpdateTime int64 `db:"update_time"`
	}

	OrderLog struct {
		Id      int32 `db:"id" auto:"yes" pk:"yes"`
		OrderId int64 `db:"order_id"`
		Type    int   `db:"type"`
		// 订单状态
		OrderState int    `db:"order_state"`
		IsSystem   int    `db:"is_system"`
		Message    string `db:"message"`
		RecordTime int64  `db:"record_time"`
	}
	OrderPromotionBind struct {
		// 编号
		Id int32 `db:"id" pk:"yes" auto:"yes"`
		// 订单号
		OrderId int32 `db:"order_id"`
		// 促销编号
		PromotionId int32 `db:"promotion_id"`
		// 促销类型
		PromotionType int `db:"promotion_type"`
		// 标题
		Title string `db:"title"`
		// 节省金额
		SaveFee float32 `db:"save_fee"`
		// 赠送积分
		PresentIntegral int `db:"present_integral"`
		// 是否应用
		IsApply int `db:"is_apply"`
		// 是否确认
		IsConfirm int `db:"is_confirm"`
	}

	// 应用到订单的优惠券
	OrderCoupon struct {
		OrderId      int32   `db:"order_id"`
		CouponId     int32   `db:"coupon_id"`
		CouponCode   string  `db:"coupon_code"`
		Fee          float32 `db:"coupon_fee"`
		Describe     string  `db:"coupon_describe"`
		SendIntegral int     `db:"send_integral"`
	}
)

func (o *OrderCoupon) Clone(coupon promotion.ICouponPromotion,
	orderId int32, orderFee float32) *OrderCoupon {
	v := coupon.GetDetailsValue()
	o.CouponCode = v.Code
	o.CouponId = v.Id
	o.OrderId = orderId
	o.Fee = coupon.GetCouponFee(orderFee)
	o.Describe = coupon.GetDescribe()
	o.SendIntegral = v.Integral
	return o
}
