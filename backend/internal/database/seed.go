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

	seedProduction(db)
}

// seedProduction 写入工艺与机台演示数据
func seedProduction(db *gorm.DB) {
	var processCount int64
	db.Model(&model.Process{}).Count(&processCount)
	if processCount == 0 {
		processes := []model.Process{
			{Code: "CLEAN", Name: "清洗", Seq: 10},
			{Code: "LITHO", Name: "光刻", Seq: 20},
			{Code: "ETCH", Name: "刻蚀", Seq: 30},
			{Code: "DEP", Name: "薄膜沉积", Seq: 40},
			{Code: "CMP", Name: "化学机械抛光", Seq: 50},
		}
		db.Create(&processes)
	}

	var machineCount int64
	db.Model(&model.Machine{}).Count(&machineCount)
	if machineCount == 0 {
		var processes []model.Process
		db.Find(&processes)
		pid := func(code string) uint {
			for _, p := range processes {
				if p.Code == code {
					return p.ID
				}
			}
			return 0
		}
		machines := []model.Machine{
			{Code: "CLEAN-01", Name: "1号清洗机", ProcessID: pid("CLEAN"), Status: model.MachineIdle},
			{Code: "CLEAN-02", Name: "2号清洗机", ProcessID: pid("CLEAN"), Status: model.MachineIdle},
			{Code: "LITHO-01", Name: "1号光刻机", ProcessID: pid("LITHO"), Status: model.MachineIdle},
			{Code: "LITHO-02", Name: "2号光刻机", ProcessID: pid("LITHO"), Status: model.MachineMaintenance, Remark: "定期保养"},
			{Code: "ETCH-01", Name: "1号刻蚀机", ProcessID: pid("ETCH"), Status: model.MachineIdle},
			{Code: "ETCH-02", Name: "2号刻蚀机", ProcessID: pid("ETCH"), Status: model.MachineIdle},
			{Code: "ETCH-03", Name: "3号刻蚀机", ProcessID: pid("ETCH"), Status: model.MachineIdle},
			{Code: "DEP-01", Name: "CVD沉积机1号", ProcessID: pid("DEP"), Status: model.MachineIdle},
			{Code: "DEP-02", Name: "PVD沉积机1号", ProcessID: pid("DEP"), Status: model.MachineFault, Remark: "射频电源异常"},
			{Code: "CMP-01", Name: "抛光机1号", ProcessID: pid("CMP"), Status: model.MachineIdle},
		}
		db.Create(&machines)
	}
}
