# Auth Service 🔐

Serviço de autenticação e gerenciamento de chaves de API do **ToggleMaster**. Este serviço é responsável por validar requisições e gerar chaves de API para acesso aos outros serviços.

## 🎯 Descrição do Serviço

O Auth Service atua como a porta de entrada (gateway) de autenticação para a arquitetura ToggleMaster. Ele:

1. Cria e gerencia chaves de API para outros serviços
2. Valida requisições recebidas de clientes e serviços internos
3. Mantém um registro de chaves ativas/inativas no PostgreSQL
4. Fornece endpoints administrativos para criar e revogar chaves
5. Expõe um endpoint `/validate` para validação de chaves

**Função crítica:** Nenhum outro serviço pode ser acessado sem uma chave de API válida criada por este serviço.

## 📦 Stack Técnico

- **Linguagem:** Go 1.21+
- **Framework:** Gin Web Framework
- **Banco de Dados:** PostgreSQL
- **Autenticação:** Bearer Token (Header Authorization)
- **Dependências principais:** github.com/lib/pq, golang.org/x/crypto

## 🚀 Como Usar

### Pré-requisitos Locais

- Go 1.21 ou superior
- PostgreSQL 12+ (instalado ou via Docker)
- Git

### Setup Local

#### 1. Clone e Navegue para o Diretório
```bash
cd Auth-Service
```

#### 2. Prepare o Banco de Dados

Crie um banco de dados PostgreSQL:
```bash
createdb auth_db
```

Execute o script de inicialização:
```bash
psql -U seu_usuario -d auth_db -f db/init.sql
```

Este script cria a tabela `api_keys` com a seguinte estrutura:
```sql
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    key VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP
);
```

#### 3. Configure as Variáveis de Ambiente
Crie um arquivo `.env` na raiz do serviço:

```env
# Banco de Dados PostgreSQL
DATABASE_URL=postgres://usuario:senha@localhost:5432/auth_db

# Ou configure individualmente:
POSTGRES_USER=togglemaster
POSTGRES_PASSWORD=seu_password_seguro
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=auth_db

# Serviço
PORT=8001

# Chave mestra para criar chaves de API (MUITO IMPORTANTE - Mude em produção!)
MASTER_KEY=admin-secreto-123

# Ambiente
ENVIRONMENT=development
```

#### 4. Instale as Dependências
```bash
go mod tidy
```

#### 5. Inicie o Serviço
```bash
go run .
```

O servidor estará disponível em `http://localhost:8001`.

### Testando Localmente

#### Health Check
```bash
curl http://localhost:8001/health
# Resposta esperada: {"status":"ok"}
```

#### Criar uma Chave de API
```bash
curl -X POST http://localhost:8001/admin/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin-secreto-123" \
  -d '{"name": "meu-primeiro-servico"}'

# Resposta esperada:
# {
#   "name": "meu-primeiro-servico",
#   "key": "tm_key_a1b2c3d4e5f6g7h8...",
#   "message": "Guarde esta chave com segurança! Você não poderá vê-la novamente."
# }
```

#### Validar uma Chave de API
```bash
curl http://localhost:8001/validate \
  -H "Authorization: Bearer tm_key_a1b2c3d4e5f6g7h8..."

# Resposta esperada: {"message":"Chave válida"}
```

#### Testar com Chave Inválida
```bash
curl http://localhost:8001/validate \
  -H "Authorization: Bearer chave-errada"

# Resposta esperada: Chave de API inválida ou inativa
```

## 🔧 Variáveis de Ambiente

### Obrigatórias
| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `POSTGRES_USER` | Usuário PostgreSQL | `togglemaster` |
| `POSTGRES_PASSWORD` | Senha PostgreSQL | `senha_forte_123` |
| `POSTGRES_HOST` | Host PostgreSQL | `localhost` |
| `POSTGRES_PORT` | Porta PostgreSQL | `5432` |
| `POSTGRES_DB` | Nome do banco de dados | `auth_db` |
| `MASTER_KEY` | Chave mestra para criar chaves de API | `seu-key-secreto` |

**Nota:** Alternativamente, use `DATABASE_URL` ao invés das variáveis individuais.

### Opcionais
| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `PORT` | Porta do servidor | `8001` |
| `ENVIRONMENT` | Ambiente (development/production) | `development` |
| `LOG_LEVEL` | Nível de log | `INFO` |
| `KEY_EXPIRATION_DAYS` | Dias até expiração de chave (0 = nunca) | `0` |

## 🔐 GitHub Secrets Necessários

Configure os seguintes secrets no GitHub para CI/CD:

```yaml
POSTGRES_USER
  Descrição: Usuário PostgreSQL
  Valor: togglemaster

POSTGRES_PASSWORD
  Descrição: Senha PostgreSQL
  Valor: <sua-senha-forte>

POSTGRES_HOST
  Descrição: Host PostgreSQL (produção)
  Valor: db.example.com

POSTGRES_PORT
  Descrição: Porta PostgreSQL
  Valor: 5432

POSTGRES_DB
  Descrição: Nome do banco de dados
  Valor: auth_db

MASTER_KEY
  Descrição: Chave mestra para criar APIs (NUNCA compartilhar)
  Valor: <seu-key-aleatorio-seguro>

DATABASE_URL
  Descrição: String de conexão completa (para staging/production)
  Valor: postgres://user:password@host:5432/auth_db

DOCKERHUB_USERNAME
  Descrição: Docker Hub username
  Valor: <seu-username>

DOCKERHUB_TOKEN
  Descrição: Docker Hub personal access token
  Valor: <seu-token>

REGISTRY_URL
  Descrição: URL do registry de container (ECR, Docker Hub, etc)
  Valor: docker.io

SONAR_TOKEN
  Descrição: Token SonarQube para análise de código
  Valor: <seu-token>
```

## 📊 Endpoints da API

| Método | Endpoint | Requer Auth | Descrição |
|--------|----------|------------|-----------|
| GET | `/health` | Não | Verifica saúde do serviço |
| POST | `/admin/keys` | Sim (MASTER_KEY) | Cria uma nova chave de API |
| GET | `/validate` | Sim (Bearer token) | Valida a chave de API |
| GET | `/admin/keys` | Sim (MASTER_KEY) | Lista todas as chaves |
| DELETE | `/admin/keys/:id` | Sim (MASTER_KEY) | Revoga uma chave |

## 🏗️ Arquitetura

```
Request → Gin Router → Handler → Database → Response
                      ↓
                   Validation
                      ↓
                  Authentication
```

## 🐛 Troubleshooting

### Problema: "postgres: could not connect"
**Solução:** Verifique se PostgreSQL está rodando e as credenciais estão corretas
```bash
psql -U seu_usuario -d auth_db -c "SELECT 1"
```

### Problema: "MASTER_KEY incorrect"
**Solução:** Verifique se está usando a chave correta no header `Authorization: Bearer`

### Problema: Chave de API não funciona
**Solução:** Verifique se a chave ainda está ativa (não foi revogada)
```bash
psql -U seu_usuario -d auth_db -c "SELECT * FROM api_keys WHERE name = 'seu-servico'"
```

## 🔄 Fluxo de Autenticação

1. **Criação de Chave:**
   - Cliente autenticado com MASTER_KEY faz requisição POST a `/admin/keys`
   - Serviço gera chave aleatória `tm_key_...`
   - Chave é armazenada em hash no PostgreSQL
   - Chave é retornada UMA VEZ ao cliente

2. **Validação de Chave:**
   - Serviço recebe header `Authorization: Bearer tm_key_...`
   - Valida a chave contra banco de dados
   - Retorna status de validade

## 📚 Recursos Adicionais

- [Go Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [ToggleMaster Architecture](../README.md)

## 👥 Suporte

Para dúvidas ou problemas, abra uma issue no repositório principal ou entre em contato com o time DevOps.
