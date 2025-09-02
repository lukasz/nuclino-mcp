# Plan Implementacji Serwera MCP dla Nuclino w Go

## 📋 Przegląd Projektu

Serwer MCP (Model Context Protocol) dla Nuclino API napisany w Go, umożliwiający integrację z Claude Desktop i innymi klientami MCP.

## 🎯 Cele Implementacji

1. Pełna obsługa Nuclino API v0
2. Zgodność z protokołem MCP
3. Komunikacja przez STDIO (stdin/stdout)
4. Obsługa wszystkich typów zasobów Nuclino
5. Efektywna obsługa błędów i rate limiting
6. Wysoka testowalność i łatwość utrzymania

## 📁 Struktura Projektu

```
nuclino-mcp-server/
├── cmd/
│   └── server/
│       └── main.go                 # Punkt wejścia aplikacji
├── internal/
│   ├── mcp/
│   │   ├── server.go               # Główna logika serwera MCP
│   │   ├── protocol.go             # Definicje protokołu MCP
│   │   ├── transport.go            # Transport STDIO
│   │   └── handlers.go             # Handlery dla metod MCP
│   ├── nuclino/
│   │   ├── client.go               # Klient HTTP dla Nuclino API
│   │   ├── types.go                # Struktury danych Nuclino
│   │   ├── auth.go                 # Autoryzacja i uwierzytelnianie
│   │   └── errors.go               # Obsługa błędów API
│   └── tools/
│       ├── registry.go             # Rejestr wszystkich narzędzi
│       ├── users.go                # Narzędzia dla użytkowników
│       ├── teams.go                # Narzędzia dla zespołów
│       ├── workspaces.go           # Narzędzia dla workspace'ów
│       ├── items.go                # Narzędzia dla items
│       ├── collections.go          # Narzędzia dla kolekcji
│       ├── fields.go               # Narzędzia dla pól
│       └── files.go                # Narzędzia dla plików
├── pkg/
│   ├── markdown/
│   │   └── converter.go            # Konwersja i walidacja Markdown
│   └── utils/
│       ├── retry.go                # Logika retry dla rate limiting
│       └── pagination.go           # Pomocniki do paginacji
├── tests/
│   ├── integration/                # Testy integracyjne
│   └── mocks/                      # Mocki dla testów
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── README.md
├── LICENSE
└── .env.example
```

## 🛠 Zalecane Biblioteki

### Podstawowe Zależności

```go
// go.mod
module github.com/yourusername/nuclino-mcp-server

go 1.21

require (
    // HTTP Client - sprawdzony, bogaty w funkcje
    github.com/go-resty/resty/v2 v2.11.0
    
    // Alternatywa: lżejszy klient HTTP z retry
    github.com/hashicorp/go-retryablehttp v0.7.5
    
    // JSON-RPC - gotowa implementacja protokołu
    github.com/ethereum/go-ethereum/rpc v1.13.0
    
    // Konfiguracja i zmienne środowiskowe
    github.com/spf13/viper v1.18.0
    github.com/joho/godotenv v1.5.1
    
    // CLI i flagi
    github.com/spf13/cobra v1.8.0
    
    // Strukturalne logowanie
    github.com/rs/zerolog v1.31.0
    
    // Walidacja struktur
    github.com/go-playground/validator/v10 v10.16.0
    
    // Rate limiting
    golang.org/x/time v0.5.0  // zawiera rate.Limiter
    
    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    
    // JSON manipulation
    github.com/tidwall/gjson v1.17.0
    github.com/tidwall/sjson v1.2.5
    
    // Markdown processing
    github.com/gomarkdown/markdown v0.0.0-20231222211730-1d6d20845b47
    
    // Schema generation/validation
    github.com/invopop/jsonschema v0.12.0
)
```

### Uzasadnienie Wyboru Bibliotek

#### **HTTP Client**
- **go-resty/resty/v2** - Bogaty w funkcje, wbudowane retry, łatwy w użyciu
- **Alternatywa: hashicorp/go-retryablehttp** - Lżejszy, automatyczne retry z exponential backoff

#### **JSON-RPC**
- **ethereum/go-ethereum/rpc** - Dojrzała implementacja JSON-RPC, sprawdzona w produkcji
- Oszczędza czas na implementacji protokołu od zera

#### **Konfiguracja**
- **spf13/viper** - Kompleksowe zarządzanie konfiguracją (env, files, flags)
- **spf13/cobra** - Profesjonalne CLI z subcommands i auto-complete

#### **Logowanie**
- **rs/zerolog** - Najszybszy strukturalny logger, zero alokacji
- **Alternatywa: uber-go/zap** - Również bardzo wydajny

#### **Testing**
- **stretchr/testify** - Bogaty zestaw asercji, suite'y testowe
- **golang/mock** - Oficjalne narzędzie do generowania mocków

#### **Markdown**
- **gomarkdown/markdown** - Pełne wsparcie CommonMark i rozszerzeń
- Konwersja Markdown ↔ HTML jeśli potrzebna

## 📝 Lista Narzędzi MCP do Implementacji

### Narzędzia Podstawowe (Priorytet 1)
1. **Items Management**
    - `nuclino_search_items` - Wyszukiwanie z pełnym wsparciem filtrów
    - `nuclino_get_item` - Pobieranie z treścią Markdown
    - `nuclino_create_item` - Tworzenie z walidacją Markdown
    - `nuclino_update_item` - Aktualizacja częściowa i pełna
    - `nuclino_delete_item` - Soft delete (kosz)
    - `nuclino_move_item` - Przenoszenie między kolekcjami

2. **Workspace Management**
    - `nuclino_list_workspaces` - Z paginacją
    - `nuclino_get_workspace` - Ze szczegółami struktury
    - `nuclino_create_workspace` - Z opcjonalnymi polami
    - `nuclino_update_workspace` - Nazwa i pola
    - `nuclino_delete_workspace` - Z potwierdzeniem

### Narzędzia Rozszerzone (Priorytet 2)
3. **Collections**
    - `nuclino_create_collection` - Tworzenie kolekcji
    - `nuclino_update_collection` - Edycja metadanych
    - `nuclino_list_collection_items` - Zawartość kolekcji
    - `nuclino_reorder_collection` - Zmiana kolejności

4. **Teams & Users**
    - `nuclino_list_teams` - Dostępne zespoły
    - `nuclino_get_team` - Szczegóły zespołu
    - `nuclino_list_team_members` - Członkowie zespołu
    - `nuclino_get_user` - Profil użytkownika
    - `nuclino_get_current_user` - Bieżący użytkownik

### Narzędzia Zaawansowane (Priorytet 3)
5. **Files & Attachments**
    - `nuclino_upload_file` - Upload z walidacją typu
    - `nuclino_get_file` - Metadane pliku
    - `nuclino_download_file` - Pobieranie zawartości
    - `nuclino_list_workspace_files` - Wszystkie pliki

6. **Fields & Metadata**
    - `nuclino_list_workspace_fields` - Pola workspace'a
    - `nuclino_create_field` - Definiowanie nowych pól
    - `nuclino_update_field` - Modyfikacja pól
    - `nuclino_delete_field` - Usuwanie pól

## 🔧 Kluczowe Komponenty Implementacji

### 1. Serwer MCP
- Użyj **JSON-RPC** biblioteki dla protokołu
- Implementuj wszystkie wymagane metody: `initialize`, `tools/list`, `tools/call`
- Obsługa STDIO z buforowaniem
- Graceful shutdown

### 2. Klient Nuclino
- Użyj **go-resty** dla łatwej obsługi HTTP
- Implementuj interfejs z metodami dla każdego endpointa
- Automatyczne retry z **exponential backoff**
- Cachowanie dla często używanych danych

### 3. Rate Limiting
- Użyj **golang.org/x/time/rate** dla token bucket
- Globalne i per-endpoint limity
- Automatyczne dostosowanie do nagłówków rate limit

### 4. Obsługa Błędów
- Własne typy błędów z kontekstem
- Mapowanie błędów Nuclino na MCP
- Strukturalne logowanie błędów

## 🧪 Strategia Testowania

### Poziomy Testów
1. **Unit Tests** (80% pokrycia)
    - Używaj **testify** dla czytelnych asercji
    - **golang/mock** dla mocków zewnętrznych serwisów
    - Table-driven tests dla parametrycznych przypadków

2. **Integration Tests**
    - Testy z prawdziwym API (opcjonalne)
    - Używaj **testcontainers-go** dla lokalnego środowiska
    - Nagrywaj i odtwarzaj requesty z **go-vcr**

3. **E2E Tests**
    - Pełne scenariusze użycia
    - Test z rzeczywistym Claude Desktop (manual)

### Narzędzia Testowe
```bash
# Generowanie mocków
mockgen -source=internal/nuclino/client.go -destination=tests/mocks/client_mock.go

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmarki
go test -bench=. -benchmem ./...
```

## 🚀 Build i Deployment

### Makefile Targets
```makefile
build:        # Kompilacja binarna
test:         # Uruchomienie testów
lint:         # golangci-lint
fmt:          # Formatowanie kodu
install:      # Instalacja zależności
docker:       # Build obrazu Docker
release:      # goreleaser dla multi-platform
```

### CI/CD Pipeline (GitHub Actions)
1. Lint i formatowanie
2. Testy jednostkowe
3. Build dla wielu platform
4. Automatyczne release z tagów

### Konfiguracja
```yaml
# config.yaml lub zmienne środowiskowe
server:
  log_level: info
  timeout: 30s

nuclino:
  api_key: ${NUCLINO_API_KEY}
  base_url: https://api.nuclino.com
  rate_limit:
    requests_per_second: 10
    burst: 20

cache:
  ttl: 5m
  max_size: 100MB
```

## 📊 Metryki i Monitoring

### Metryki do Śledzenia
- Liczba wywołań per narzędzie
- Czas odpowiedzi per endpoint
- Rate limit hits
- Błędy per typ
- Cache hit ratio

### Implementacja
- Użyj **prometheus/client_golang** dla metryk
- Opcjonalnie **OpenTelemetry** dla distributed tracing

## 🔒 Bezpieczeństwo

### Best Practices
1. **Secrets Management**
    - Nigdy nie loguj API key
    - Użyj **hashicorp/vault** dla produkcji

2. **Input Validation**
    - **go-playground/validator** dla struktur
    - Sanityzacja Markdown przed wysłaniem

3. **Security Headers**
    - Timeout na wszystkie requesty
    - Limit rozmiaru payloadu

## 📚 Dokumentacja

### Struktura README
1. **Quick Start** - 5 minut do uruchomienia
2. **Installation** - Wszystkie metody instalacji
3. **Configuration** - Pełna lista opcji
4. **Tools Reference** - Auto-generowana z kodu
5. **Examples** - Rzeczywiste use cases
6. **Troubleshooting** - Częste problemy
7. **Contributing** - Jak pomóc w rozwoju

### Generowanie Dokumentacji
- Użyj **godoc** dla dokumentacji API
- **swagger** dla REST endpoints (jeśli dodasz HTTP server)
- Przykłady w `examples/` directory

## 🎯 Kamienie Milowe Implementacji

### Faza 1: Fundament (2-3 dni)
- [ ] Setup projektu z wszystkimi zależnościami
- [ ] Podstawowa struktura MCP server
- [ ] Klient Nuclino z autoryzacją
- [ ] Pierwsze działające narzędzie (get_item)

### Faza 2: Core Features (3-4 dni)
- [ ] Wszystkie narzędzia Items
- [ ] Workspace management
- [ ] Search z filtrami
- [ ] Unit testy dla core

### Faza 3: Extended Features (2-3 dni)
- [ ] Collections support
- [ ] Teams i Users
- [ ] Files handling
- [ ] Integration testy

### Faza 4: Polish (2-3 dni)
- [ ] Rate limiting i retry
- [ ] Caching layer
- [ ] Kompleksowa dokumentacja
- [ ] CI/CD pipeline
- [ ] Release preparation

## 🤝 Wskazówki dla Claude Code

### Kolejność Implementacji
1. **Start Simple** - Zacznij od jednego działającego narzędzia
2. **Vertical Slice** - Pełny stack dla jednej funkcji przed dodaniem kolejnych
3. **Test Early** - Pisz testy wraz z kodem
4. **Refactor Often** - Wydzielaj wspólny kod do utils

### Struktura Sesji
```bash
# Sesja 1: Bootstrap
"Stwórz podstawową strukturę projektu MCP server dla Nuclino w Go z zalecanymi bibliotekami"

# Sesja 2: Core Implementation
"Zaimplementuj klienta Nuclino i podstawowe narzędzia dla Items"

# Sesja 3: Extended Features
"Dodaj obsługę Collections, Workspaces i Search"

# Sesja 4: Production Ready
"Dodaj rate limiting, caching, comprehensive error handling i testy"

# Sesja 5: Documentation
"Stwórz kompletną dokumentację i przykłady użycia"
```

### Dobre Praktyki Go
- Używaj interfejsów dla testowalności
- Embedded structs dla kompozycji
- Context dla timeout i cancellation
- Defer dla cleanup
- Named returns sparingly
- Error wrapping z `fmt.Errorf("%w")`

## 🔗 Przydatne Zasoby

### Dokumentacja
- [Nuclino API Docs](https://help.nuclino.com/d3a29686-api)
- [MCP Protocol Spec](https://modelcontextprotocol.io/docs)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

### Przykładowe Implementacje
- [MCP Server Examples](https://github.com/modelcontextprotocol/servers)
- [Go REST API Best Practices](https://github.com/golang-standards/project-layout)
- [Production-Ready Go Service](https://github.com/ardanlabs/service)

### Narzędzia Deweloperskie
- [golangci-lint](https://golangci-lint.run/) - Meta linter
- [goreleaser](https://goreleaser.com/) - Release automation
- [air](https://github.com/cosmtrek/air) - Live reload
- [delve](https://github.com/go-delve/delve) - Debugger

## ✅ Checklist Przed Release

- [ ] Wszystkie testy przechodzą
- [ ] Dokumentacja kompletna
- [ ] Przykłady dla każdego narzędzia
- [ ] Linting bez błędów
- [ ] Security scan (gosec)
- [ ] Performance profiling
- [ ] Multi-platform builds
- [ ] Docker image
- [ ] Version tagging
- [ ] CHANGELOG updated