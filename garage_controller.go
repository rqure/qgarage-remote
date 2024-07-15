package main

import qdb "github.com/rqure/qdb/src"

type EventType int

type IEvent interface {
	GetType() EventType
}

type GarageController struct {
	db     qdb.IDatabase
	events chan IEvent
}

func NewGarageController(db qdb.IDatabase) *GarageController {
	return &GarageController{
		db: db,
	}
}

func (gc *GarageController) Init() {

}

func (gc *GarageController) Deinit() {

}

func (gc *GarageController) DoWork() {

}
