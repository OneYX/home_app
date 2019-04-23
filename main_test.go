package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	fmt.Println(strconv.FormatInt(1<<6-1, 2))
}

func TestTime(t *testing.T) {
	fmt.Println()
}

//func (d Person) MarshalJSON() ([]byte, error) {
//	type Alias Person
//	return json.Marshal(&struct {
//		Person
//		CreateTime string `json:"create_time"`
//	}{
//		Person:     d,
//		CreateTime: d.CreateTime.Format("2006/01/02 15:04:05"),
//	})
//}

type Person struct {
	CreateTime time.Time `json:"create_time"`
}

type Persons struct {
	Person
	CreateTime string `json:"create_time"`
}

func TestAlias(t *testing.T) {
	p := Persons{Person{}, time.Now().Format("2006/01/02 15:04:05")}
	json.Marshal(p)
}
