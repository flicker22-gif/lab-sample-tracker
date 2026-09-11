package database

import (
	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

// seed 写入初始用户与位置（仅首次启动）
func seed(db *gorm.DB) {
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		users := []model.User{
			{Username: "zhangsan", FullName: "张三"},
			{Username: "lisi", FullName: "李四"},
			{Username: "wangwu", FullName: "王五"},
		}
		db.Create(&users)
	}

	var locCount int64
	db.Model(&model.Location{}).Count(&locCount)
	if locCount == 0 {
		locs := []model.Location{
			{Name: "冰箱A2-3层", Type: model.LocFridge, Building: "2号楼", Room: "201", Temperature: -20},
			{Name: "冰箱A2-5层", Type: model.LocFridge, Building: "2号楼", Room: "201", Temperature: -20},
			{Name: "超低温冰箱B1", Type: model.LocFridge, Building: "2号楼", Room: "203", Temperature: -80},
			{Name: "冷藏柜C3", Type: model.LocFridge, Building: "2号楼", Room: "205", Temperature: 4},
			{Name: "HPLC-01", Type: model.LocDevice, Building: "2号楼", Room: "302"},
			{Name: "GC-MS-02", Type: model.LocDevice, Building: "2号楼", Room: "302"},
			{Name: "ICP-MS-01", Type: model.LocDevice, Building: "2号楼", Room: "305"},
			{Name: "收样台", Type: model.LocBench, Building: "2号楼", Room: "101"},
			{Name: "销毁点", Type: model.LocDiscard, Building: "2号楼", Room: "101"},
		}
		db.Create(&locs)
	}
}
