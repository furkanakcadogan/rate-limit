package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
)

func main() {
	pgConnStr := os.Getenv("DB_SOURCE")

	postgres_db, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	defer postgres_db.Close()

	// Kullanıcı bilgilerini burada belirleyin
	clientID := "yeni-kullanici"
	rateLimit := int32(200)
	refillInterval := int32(7200)

	// Call the CreateNewUser function from the same package
	_, err = CreateNewUser(context.Background(), postgres_db, clientID, rateLimit, refillInterval)
	if err != nil {
		log.Fatalf("Kullanıcı oluşturma hatası: %v", err)
	}

	fmt.Println("Yeni kullanıcı başarıyla oluşturuldu!")
}
