package models

import (
	"testing"
)

func TestInitDB(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Errorf("InitDB failed: %v", err)
	}

	if db == nil {
		t.Error("DB should not be nil after InitDB")
	}
}

func TestGetDB(t *testing.T) {
	InitDB()

	database := GetDB()
	if database == nil {
		t.Error("GetDB should not return nil")
	}
}

func TestCloseDB(t *testing.T) {
	InitDB()

	err := CloseDB()
	if err != nil {
		t.Errorf("CloseDB failed: %v", err)
	}
}
