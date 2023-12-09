# Resmi Go imajını kullanın
FROM golang:1.17

# Çalışma dizini oluşturun
WORKDIR /app

# Uygulama dosyalarını kopyalayın
COPY cmd/app/app.go /app/

# Uygulamayı derleyin
RUN go build app.go

# Eğer uygulama bağımlılıkları varsa onları ekleyin (örneğin go mod dosyası)
COPY go.mod go.sum /app/
RUN go mod download

# Uygulama çalıştırılacak komut
CMD ["./app"]
