package model

// DeviceCreate 创建设备请求
type DeviceCreate struct {
	VIN         string `json:"vin" binding:"required" gorm:"type:varchar(20);not null;comment:车架号"`
	CarPlate    string `json:"car_plate" binding:"required" gorm:"type:varchar(20);uniqueIndex;not null;comment:车牌号"`
	Power       int    `json:"power" binding:"required" gorm:"type:int;not null;comment:功率"`
	Speed       int    `json:"speed" gorm:"type:int;index;not null;comment:速度"`
	Status      int    `json:"status" gorm:"type:tinyint;default:0;comment:状态 0-离线 1-在线"`
	Lat         string `json:"lat" gorm:"type:varchar(32);comment:纬度"`
	Lon         string `json:"lon" gorm:"type:varchar(32);comment:经度"`
	Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
}

// DeviceUpdate 更新设备请求
type DeviceUpdate struct {
	VIN         *string `json:"vin" gorm:"type:varchar(20);not null;comment:车架号"`
	CarPlate    *string `json:"car_plate" gorm:"type:varchar(20);uniqueIndex;not null;comment:车牌号"`
	Power       *int    `json:"power" gorm:"type:int;not null;comment:功率"`
	Speed       *int    `json:"speed" gorm:"type:int;index;not null;comment:速度"`
	Status      *int    `json:"status" gorm:"type:tinyint;default:0;comment:状态 0-离线 1-在线"`
	Lat         *string `json:"lat" gorm:"type:varchar(32);comment:纬度"`
	Lon         *string `json:"lon" gorm:"type:varchar(32);comment:经度"`
	Description *string `json:"description" gorm:"type:varchar(255);comment:描述"`
}

// Device IoT 设备模型示例
type Device struct {
	BaseModel
	VIN         string `json:"vin" gorm:"type:varchar(20);not null;comment:车架号"`
	CarPlate    string `json:"car_plate" gorm:"type:varchar(20);uniqueIndex;not null;comment:车牌号"`
	Power       int    `json:"power" gorm:"type:int;not null;comment:功率"`
	Speed       int    `json:"speed" gorm:"type:int;index;not null;comment:速度"`
	Status      int    `json:"status" gorm:"type:tinyint;default:0;comment:状态 0-离线 1-在线"`
	Lat         string `json:"lat" gorm:"type:varchar(32);comment:纬度"`
	Lon         string `json:"lon" gorm:"type:varchar(32);comment:经度"`
	Description string `json:"description" gorm:"type:varchar(255);comment:描述"`
}

func (Device) TableName() string {
	return "devices"
}
