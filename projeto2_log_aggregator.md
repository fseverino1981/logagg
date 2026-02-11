# 🚀 Projeto 2: Log Aggregator

## 📋 Descrição

Uma ferramenta CLI que monitora múltiplos arquivos de log simultaneamente, agregando as linhas em uma única saída. Usa o padrão **Fan-In** para combinar streams de diferentes fontes.

---

## 🎯 Funcionalidades

| Funcionalidade | Descrição |
|----------------|-----------|
| Monitorar múltiplos arquivos | `./logagg --files app.log,error.log,access.log` |
| Prefixar origem | Cada linha mostra de qual arquivo veio |
| Filtrar por padrão | `./logagg --files app.log --filter "ERROR"` |
| Modo tail | Fica escutando novas linhas (como `tail -f`) |
| Graceful shutdown | Encerra limpo com Ctrl+C |

---

## 📖 Exemplos de uso

```bash
# Monitora um arquivo
./logagg --files app.log

# Monitora múltiplos arquivos
./logagg --files app.log,error.log,access.log

# Filtra apenas linhas com ERROR
./logagg --files app.log --filter "ERROR"

# Modo tail (fica escutando)
./logagg --files app.log --tail
```

**Saída esperada:**
```
[app.log] 2024-01-15 10:23:45 INFO Starting application
[error.log] 2024-01-15 10:23:46 ERROR Connection refused
[app.log] 2024-01-15 10:23:47 INFO Retrying...
```

---

## 🏗️ Arquitetura sugerida

```
logagg/
├── cmd/
│   └── root.go
├── internal/
│   ├── reader/
│   │   ├── reader.go       # Lê linhas de um arquivo (generator)
│   │   └── validator.go    # Valida se arquivo existe
│   ├── aggregator/
│   │   └── aggregator.go   # Fan-in: combina múltiplos readers
│   └── filter/
│       └── filter.go       # Filtra linhas por padrão
├── main.go
├── go.mod
├── Makefile
└── README.md
```

---

## 🔧 Padrões de concorrência

| Padrão | Onde usar |
|--------|-----------|
| **Generator** | `reader.go` — cada arquivo retorna `<-chan string` |
| **Fan-In** | `aggregator.go` — combina múltiplos channels em um |
| **Pipeline** | Filter recebe channel, retorna channel filtrado |
| **Graceful shutdown** | `context.Context` para cancelar todas as goroutines |

---

## ⚙️ Requisitos técnicos

1. **Generator por arquivo**: Cada arquivo tem sua goroutine lendo linhas
2. **Fan-In**: Função que recebe `[]<-chan string` e retorna `<-chan string`
3. **Context para cancelamento**: Ctrl+C cancela todas as goroutines
4. **WaitGroup**: Garantir que todas as goroutines terminaram
5. **Testes**: Mínimo 70% de cobertura
6. **README**: Documentação completa com decisões técnicas

---

## 💡 Dicas de implementação

### Reader (Generator)

```go
func ReadLines(ctx context.Context, filename string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        // Abre arquivo, lê linha por linha
        // Verifica ctx.Done() para cancelamento
    }()
    return out
}
```

### Aggregator (Fan-In)

```go
func Aggregate(ctx context.Context, channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup
    
    // Para cada channel de entrada, cria goroutine que repassa para out
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

### Validator

```go
func ValidateFile(filename string) error {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return errors.New("arquivo não encontrado: " + filename)
    }
    if info.IsDir() {
        return errors.New("caminho é um diretório, não arquivo: " + filename)
    }
    return nil
}
```

### Graceful shutdown

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
```

---

## 📅 Prazo sugerido

1-2 semanas. Esse projeto é mais complexo que o anterior.

---

## ✅ Critérios de avaliação

| Critério | Peso |
|----------|------|
| Funciona corretamente | 25% |
| Padrões de concorrência (generator, fan-in) | 25% |
| Graceful shutdown com context | 15% |
| Código limpo e organizado | 15% |
| Testes | 10% |
| README e documentação | 10% |

---

## 🔜 Conceitos que você vai praticar

- Generator (já conhece)
- Fan-In (já conhece)
- Pipeline
- `context.Context` para cancelamento
- `sync.WaitGroup`
- Leitura de arquivos
- `signal.NotifyContext`

---

## 📚 Referências

- [Go by Example: Reading Files](https://gobyexample.com/reading-files)
- [Go by Example: Context](https://gobyexample.com/context)
- [Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines)
