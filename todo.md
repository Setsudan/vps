# Security Hardening TODO

> Work through the checklist top‑to‑bottom.  
> Tick each box when complete and remove the item if you adopt a different solution.

---

## 1 Secrets & Credentials

- [👌] **Fail‑fast on missing `JWT_SECRET`**

  ```go
  // utils/env.go
  package utils

  import (
      "log"
      "os"
  )

  // mustEnv returns the value or terminates the app.
  func MustEnv(key string) string {
      v := os.Getenv(key)
      if v == "" {
          log.Fatalf("%s must be set", key)
      }
      return v
  }
  ```

  ```go
  // init_server.go
  var jwtSecret = []byte(utils.MustEnv("JWT_SECRET"))
  ```

- [👌] **Remove default DB credentials**  
  Update `.env.example` to leave them blank.

  ```dotenv
  POSTGRES_USER=
  POSTGRES_PASSWORD=
  ```

- [👌] **Load secrets from a manager or Docker secrets**

  ```yaml
  services:
    backend:
      secrets:
        - jwt_secret
  secrets:
    jwt_secret:
      external: true
  ```

---

## 2 · Network Hardening

- [👌] **Keep internal services private**

  ```yaml
  services:
    redis:
      networks: [internal]
      expose: ["6379"]
    postgres:
      networks: [internal]
      expose: ["5432"]
  networks:
    internal:
      internal: true
  ```

- [ ] **Remove Traefik labels that publish Redis/Postgres/SeaweedFS**

---

## 3 · Traefik Dashboard

- [👌] **Turn off insecure API**

  ```yaml
  command:
    - "--api.dashboard=true"
    - "--api.insecure=false"
  ```

- [👌] **Protect dashboard with TLS + basic‑auth**

  ```yaml
  labels:
    - traefik.http.routers.traefik.rule=Host(`traefik.example.com`)
    - traefik.http.routers.traefik.entrypoints=websecure
    - traefik.http.routers.traefik.tls.certresolver=letsencrypt
    - traefik.http.routers.traefik.middlewares=auth
    - traefik.http.middlewares.auth.basicauth.users=admin:$$apr1$$<hash>
  ```

---

## 4 · TLS Everywhere

- [👌] **Postgres DSN**

  ```go
  dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require",
      host, user, pass, db)
  ```

- [👌] **SeaweedFS S3 Client**

  ```go
  client, _ := minio.New(endpoint, &minio.Options{
      Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
      Secure: true,
  })
  ```

- [👌] **Redis** - enable TLS or put behind stunnel/Traefik `websecure`.

---

## 5 · WebSockets

- [👌] **Validate `Origin`**

  ```go
  upgrader := websocket.Upgrader{
      CheckOrigin: func(r *http.Request) bool {
          return r.Header.Get("Origin") == "https://app.example.com"
      },
  }
  ```

- [☠️] **Move JWT from query‑string to `Authorization` header**

  ```js
  // client
  const socket = new WebSocket("wss://api.example.com/ws", ["jwt", token]);
  ```

  ```go
  // server
  token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
  ```

---

## 6 · Observability Endpoints

- [ ] **Restrict `/metrics`**

  ```go
  r := gin.New()
  protected := r.Group("/", gin.BasicAuth(gin.Accounts{"prom": "<pwd>"}))
  protected.GET("/metrics", gin.WrapH(promhttp.Handler()))
  ```

---

## 7 · Containers & Images

- [👌] **Run as non‑root**

  ```dockerfile
  FROM golang:1.23.8-alpine AS build
  # build steps…

  RUN adduser -D appuser
  USER appuser
  ```

- [👌] **Pin image by digest**

  ```yaml
  image: golang@sha256:<digest>
  ```

- [❌] **Set resource limits**

  ```yaml
  deploy:
    resources:
      limits:
        cpus: "1"
        memory: 512M
  ```

---

## 8 · Migrations

- [❌] **Remove in‑app `AutoMigrate`**

  ```diff
  - db.AutoMigrate(&models.User{}, &models.Group{})
  ```

- [ ] **Run migrations in CI/CD**

  ```makefile
  migrate:
    golang-migrate -database $(DB_URL) -path migrations up
  ```

---

### Done?

When every item is ticked, commit this file as proof of the hardening work.
