# Estágio 1: "Builder"
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 1. Copia o go.mod
COPY go.mod ./

# 2. Copia TODO o código-fonte (.go) imediatamente
# O 'tidy' precisa ler os imports destes arquivos
COPY . .

# 3. AGORA sim, roda o 'tidy'
# Ele vai ler os .go, ver os imports, e CRIAR O go.sum
RUN go mod tidy

# 4. Instala o orchestrion (necessário para o comando abaixo funcionar)
RUN go install github.com/DataDog/orchestrion@latest

# 5. Fixa a versão do orchestrion no go.mod (necessário a partir da v1.x)
RUN orchestrion pin

# 6. Compila a aplicação
# O 'go build' agora usará o go.mod e o go.sum criados pelo 'tidy'
#RUN CGO_ENABLED=0 go build -o auth-service -ldflags "-s -w" .
RUN CGO_ENABLED=0 orchestrion go build -o auth-service -ldflags "-s -w" .

# -----------------------------------------------------------------
# Estágio 2: "Final" (Inalterado)
# -----------------------------------------------------------------
FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/auth-service .
EXPOSE 8001
CMD ["./auth-service"]
