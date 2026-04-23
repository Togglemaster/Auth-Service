package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

var tracer = otel.Tracer("auth-service")

// App struct (para injeção de dependência)
type App struct {
	DB        *sql.DB
	MasterKey string
}

func initTracer() func() {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("auth-service"))

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(resources),
		trace.WithBatcher(exporter))

	otel.SetTracerProvider(traceProvider)

	return func() {
		err := traceProvider.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	cleanup := initTracer()
	defer cleanup()
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(context.Background(), "main")

	// Carrega o .env para desenvolvimento local. Em produção, isso não fará nada.
	_ = godotenv.Load()

	// --- Configuração ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001" // Porta padrão
	}

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		// Caso não exista, monta a URL a partir das variáveis individuais
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		host := os.Getenv("POSTGRES_HOST")
		dbPort := os.Getenv("POSTGRES_PORT")
		dbName := os.Getenv("POSTGRES_DB")

		if user == "" || password == "" || host == "" || dbPort == "" || dbName == "" {
			log.Fatal("Variáveis de ambiente insuficientes para montar a conexão com o banco de dados")
		}

		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, dbPort, dbName)
	}

	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		log.Fatal("MASTER_KEY deve ser definida")
	}

	// --- Conexão com o Banco ---
	db, err := connectDB(databaseURL, ctx)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v, na url %s", err, databaseURL)
	}
	defer db.Close()

	app := &App{
		DB:        db,
		MasterKey: masterKey,
	}

	// --- Rotas da API ---
	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthHandler)

	// Endpoint público para validar uma chave
	mux.HandleFunc("/validate", app.validateKeyHandler)

	// Endpoints de "admin" para criar/gerenciar chaves
	// Eles são protegidos pelo middleware de autenticação
	mux.Handle("/admin/keys", app.masterKeyAuthMiddleware(http.HandlerFunc(app.createKeyHandler)))

	log.Printf("Serviço de Autenticação (Go) rodando na porta %s", port)
	defer span.End()
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// connectDB inicializa e testa a conexão com o PostgreSQL
func connectDB(databaseURL string, ctx context.Context) (*sql.DB, error) {
	ctx, span := tracer.Start(ctx, "connectDB")
	defer span.End()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	log.Println("Conectado ao PostgreSQL com sucesso!")
	return db, nil
}
