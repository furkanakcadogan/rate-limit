# Resmi Go imajını kullanın
#docker build -t rate-limiter-app .
#docker run rate-limiter-app

FROM golang:1.21

# Çalışma dizini oluşturun
WORKDIR /app

# Uygulama dosyalarını kopyalayın
COPY . /app/

# Uygulamayı derleyin
RUN go build cmd/app/app.go

# Uygulama çalıştırılacak komut
CMD ["./app"]
