# PostgreSQL (pgx) Docker Setup Guide

## 1. Run PostgreSQL Container

```bash
docker run --name <container-name> \
  -e POSTGRES_PASSWORD=<password> \
  -p 5431:5432 \
  -d postgres
```

**Flags explained:**

* `-e` : Environment variable
* `-p` : Port mapping
* `-d` : Detached mode
* `postgres` : Docker image used for the container

---

## 2. List Containers

```bash
docker ps
```

or

```bash
docker ps -a
```

**Flag:**

* `-a` : Show all containers (running + stopped)

---

## 3. TablePlus Configuration

| Field          | Value                                       |
| -------------- | ------------------------------------------- |
| **Name**       | `<container-name>` (or anything you prefer) |
| **Connection** | PostgreSQL                                  |
| **Host**       | `localhost`                                 |
| **Port**       | `5431`                                      |
| **User**       | `postgres`                                  |
| **Password**   | `<password>`                                |
| **Database**   | `postgres`                                  |

---

## 4. Connect via psql (Terminal Access)

### Option A — Through Docker

```bash
docker exec -it <container-name> psql -U postgres
```

### Option B — Using a connection string

```bash
psql "<connection-string>"
```

**Flags explained:**

* `-it` : Interactive terminal session
* `-U`  : Username

---

## 5. Stop the Container

```bash
docker stop <container-name>
```

---

## 6. Remove the Container

```bash
docker rm <container-name>
```

*(Use up migrations to recreate tables after removal.)*

---

## 7. Persisting Data Using Volumes

### Create a Docker volume

```bash
docker volume create <container-name-data>
```

### Run container with volume

```bash
docker run --name <container-name> \
  -e POSTGRES_PASSWORD=<password> \
  -p 5431:5432 \
  -v <container-name-data>:<filepath> \
  -d postgres
```

**Why volumes?**
Resetting containers clears internal data. Volumes persist data to disk so your database survives container restarts or recreation.

---

## 8. Connection String Format

```
protocol://username:password@host:port/database
```

Example values:

* **protocol:** `postgres`