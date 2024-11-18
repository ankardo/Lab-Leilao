# Lab-Leilão

## Pré-requisitos

- Docker
- Go

## Comandos necessários para executar o sistema

### Execução inicial

```bash
docker compose up --build -d
```

### Execuções futuras

```bash
docker compose up -d
```

## Como usar

- Subindo a aplicação com os comandos acima, o teste de integração será executado automaticamente.
- Para verificar os resultados do teste de integração, você pode rodar os seguintes comandos:

```bash
docker logs auction
docker logs integration-tests
```

- Você também pode consultar diretamente o banco de dados para verificar que um leilão foi criado e fechado automaticamente de acordo com o tempo definido no arquivo `.env`.

