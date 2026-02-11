.PHONY: build run test clean

build:
	go build -o logagg .

run: build
	./logagg

test:
	go test ./... -cover -v

clean:
	rm -f logagg

help:
	@echo "Comandos disponíveis:"
	@echo "  make build  - Compila o projeto"
	@echo "  make run    - Compila e executa"
	@echo "  make test   - Roda os testes com cobertura"
	@echo "  make clean  - Remove o binário"