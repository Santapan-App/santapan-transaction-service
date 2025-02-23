package main

import (
	"log"
	"os"
	"os/signal"
	"santapan_transaction_service/cart"
	postgresCommands "santapan_transaction_service/internal/repository/postgres/commands"
	postgresQueries "santapan_transaction_service/internal/repository/postgres/queries"
	"santapan_transaction_service/internal/rest"
	"santapan_transaction_service/item"
	pkgEcho "santapan_transaction_service/pkg/echo"
	"santapan_transaction_service/pkg/sql"
	"santapan_transaction_service/transaction"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import the postgres driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Import the file source driver

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

const (
	defaultTimeout = 30
	defaultAddress = ":9091"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	conn := sql.Setup()
	defer sql.Close(conn)

	cartQueryRepo := postgresQueries.NewPostgresCartQueryRepository(conn)
	cartCommandRepo := postgresCommands.NewPostgresCartCommandRepository(conn)

	itemQueryRepo := postgresQueries.NewPostgresItemQueryRepository(conn)
	itemCommandRepo := postgresCommands.NewPostgresItemCommandRepository(conn)

	transactionQueryRepo := postgresQueries.NewPostgresTransactionQueryRepository(conn)
	transactionCommandRepo := postgresCommands.NewPostgresTransactionCommandRepository(conn)

	// Service
	cartService := cart.NewService(cartQueryRepo, cartCommandRepo)
	itemService := item.NewService(itemQueryRepo, itemCommandRepo)
	transactionService := transaction.NewService(transactionQueryRepo, transactionCommandRepo)
	e := pkgEcho.Setup()

	rest.NewCartHandler(e, cartService, itemService)
	rest.NewTransactionHandler(e, transactionService, cartService)

	go func() {
		pkgEcho.Start(e)
	}()

	// Channel to listen for termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-quit

	pkgEcho.Shutdown(e, defaultTimeout)
}
