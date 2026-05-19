package dao

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Title        string  `gorm:"column:title;type:varchar(200);not null;default:''" json:"title" comment:"商品标题"`
	Description  string  `gorm:"column:description;type:text;not null;default:''" json:"description" comment:"商品描述"`
	Price        float64 `gorm:"column:price;type:decimal(10,2);not null;default:0" json:"price" comment:"价格"`
	Images       string  `gorm:"column:images;type:varchar(1000);not null;default:''" json:"images" comment:"图片URL列表,逗号分隔"`
	Location     string  `gorm:"column:location;type:varchar(100);not null;default:''" json:"location" comment:"交易地点"`
	Status       int     `gorm:"column:status;type:int;not null;default:0" json:"status" comment:"状态:0在售,1已售出,2已下架"`
	UserId       uint    `gorm:"column:user_id;type:int;not null;default:0" json:"user_id" comment:"发布者ID"`
	BuyUid       uint    `gorm:"column:buy_uid;type:int;not null;default:0" json:"buy_uid" comment:"购买者ID,0表示未售出"`
	ContactType  string  `gorm:"column:contact_type;type:varchar(10);not null;default:''" json:"contact_type" comment:"联系方式类型:phone/wechat/qq"`
	ContactValue string  `gorm:"column:contact_value;type:varchar(100);not null;default:''" json:"contact_value" comment:"联系方式值"`
}

func (Product) TableName() string {
	return "product"
}
