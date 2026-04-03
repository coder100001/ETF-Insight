package main

import (
	"fmt"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

func main() {
	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		fmt.Printf("Failed to initialize