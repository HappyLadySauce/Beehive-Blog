package filedrivers

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
)

func TestValidateDeps(t *testing.T) {
	tests := []struct {
		name       string
		svc        *svc.ServiceContext
		wantSubstr string
	}{
		{name: "nil context", svc: nil, wantSubstr: "service context is nil"},
		{name: "nil db", svc: &svc.ServiceContext{}, wantSubstr: "database handle is nil"},
		{
			name:       "nil driver store",
			svc:        &svc.ServiceContext{DB: newFileDriversTestDB(t)},
			wantSubstr: "driver store is nil",
		},
		{
			name: "nil registry",
			svc: &svc.ServiceContext{
				DB:          newFileDriversTestDB(t),
				DriverStore: driver.NewStore(newFileDriversTestDB(t)),
			},
			wantSubstr: "driver registry is nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeps(tt.svc)
			if err == nil {
				t.Fatal("validateDeps() = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Fatalf("validateDeps() = %q, want substring %q", err.Error(), tt.wantSubstr)
			}
		})
	}
}

func TestParseIDParam(t *testing.T) {
	id, err := parseIDParam("42")
	if err != nil || id != 42 {
		t.Fatalf("parseIDParam(42) = (%d, %v)", id, err)
	}
	if _, err := parseIDParam("not-a-number"); err == nil {
		t.Fatal("parseIDParam(invalid) = nil, want error")
	}
}

func TestNewFileDriversControllerSuccess(t *testing.T) {
	db := newFileDriversTestDB(t)
	svcCtx := &svc.ServiceContext{
		DB:             db,
		DriverStore:    driver.NewStore(db),
		DriverRegistry: driver.NewDriverRegistry(),
	}
	h, err := NewFileDriversController(svcCtx)
	if err != nil {
		t.Fatalf("NewFileDriversController: %v", err)
	}
	if h == nil || h.driverStore == nil || h.driverRegistry == nil {
		t.Fatal("controller or dependencies are nil")
	}
}

func newFileDriversTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db
}
