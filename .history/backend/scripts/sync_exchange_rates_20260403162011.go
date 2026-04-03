package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// FrankfurterAPIResponse Frankfurter API响应
type FrankfurterAPIResponse struct {
