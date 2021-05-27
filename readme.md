# Vault-Observe
> An application to send Vault Audit Events to a Tracing provider


## Building

```bash
go build
```

## Usage

Assuming you have a Vault instance running on the same machine, start a copy of `vault-observe`:

```bash
# honeycomb
./vault-observe --socket-path /opt/vault/observe.sock --honeycomb

# zipkin
./vault-observe --socket-path /opt/vault/observe.sock --zipkin
```

In another terminal, use the Vault CLI to configure an audit backend:

```bash
vault audit enable socket address=/opt/vault/observe.sock socket_type=unix
```

## Testing

Run docker-compose to get a Vault and a Zipkin container:

```bash
docker-compose up -d
./vault-observe --zipkin
```

In another terminal, enable the audit backend:

```bash
export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="vault"

vault audit enable socket address=/sockets/observe.sock socket_type=unix
```

Make some requests to see the results in the Zipkin UI ([http://localhost:9411/zipkin](http://localhost:9411/zipkin)):

```bash
vault secrets enable -version=2 kv
vault kv put /secrets/test name=andy

vault kv get /secrets/test
vault kv get /secrets/test
vault kv get /secrets/test
vault kv get /secrets/test
```