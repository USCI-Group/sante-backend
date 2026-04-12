This is a guide to install and run postgres in docker in the server.

## Pull postgres image
```bash
docker pull postgres
```

## Create a volume for postgres
```bash
docker volume create postgres_data
```

## Run postgres container, with the volume and database
```bash
docker run --name postgres-container \
  -e POSTGRES_USER=myuser \
  -e POSTGRES_PASSWORD=mypassword \
  -e POSTGRES_DB=mydb \
  -v postgres_data:/var/lib/postgresql/data \
  -p 5432:5432 \
  -d postgres
```

Don't forget allow port 5432 in the ec2 security group.