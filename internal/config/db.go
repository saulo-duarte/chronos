package config

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

func Connect(ctx context.Context, dsn string) error {
	var err error
	once.Do(func() {
		log := WithContext(ctx)

		sqlDB, errOpen := gorm.Open(postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}), &gorm.Config{})
		if errOpen != nil {
			log.WithError(errOpen).Errorf("Falha inicial ao abrir a conexão com o banco de dados")
			err = errOpen
			return
		}

		var db *sql.DB
		db, err = sqlDB.DB()
		if err != nil {
			log.WithError(err).Errorf("Falha ao obter a instância sql.DB")
			return
		}

		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		const maxRetries = 5
		const retryDelay = 2 * time.Second

		for i := 0; i < maxRetries; i++ {
			if pingErr := db.Ping(); pingErr == nil {
				DB = sqlDB
				log.Info("Conexão com o banco de dados estabelecida com sucesso")
				return
			} else {
				log.WithField("tentativa", i+1).WithError(pingErr).Warn("Falha ao conectar com o banco de dados. Tentando novamente...")
				time.Sleep(retryDelay)
				err = pingErr
			}
		}

		log.WithError(err).Error("Número máximo de tentativas de conexão excedido")
		err = fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
	})

	return err
}
