package database

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New() // Мок бд
	if err != nil {
		t.Fatal(err)
	}
	// конектимся к моку
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// функция для закрытия мока бд
	cleanup := func() {
		db.Close()
	}

	return gormDB, mock, cleanup
}
func TestSaveOrder_Success(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t) // Gorm с мок бд
	defer cleanup()

	repo := NewGormDatabase(gormDB, 1, 0)
	// ожидаем цепочку: начало транзакции -> оперцию к бд -> успешный результат -> коммит
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "orders"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	order := &Order{OrderUID: "123"}

	err := repo.SaveOrder(order)
	if err != nil {
		t.Fatal(err)
	}

	// проверяем что все по ожидаемое произошло
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestSaveOrder_RollbackOnError(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewGormDatabase(gormDB, 1, 0)
	// ожидаем ошибку и откат
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "orders"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	order := &Order{}
	err := repo.SaveOrder(order)
	if err == nil {
		t.Fatal("expected error")
	}

	// проверяем что все по ожидаемое произошло
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestGetOrder_NotFound_NoRetry(t *testing.T) {
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewGormDatabase(gormDB, 1, 0)
	// запись не найдена
	mock.ExpectQuery(`SELECT (.+)FROM "orders"`).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.GetOrder("uid-x")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
