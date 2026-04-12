## Migration script

In the project main folder...
To run Migration script, go into database dir, and initialize the database.

```bash
cd database
./scripts/generate-migration init
```

Run this to check if the connection is successful

```bash
encore db conn-uri sante
# or
# encore db conn-uri sante --shadow
```

## set environment variable

1. Set environment variable to local

```bash
export ENV=local
```

2. Add .env file to the root directory
   Please ask for .env file from the team leader.

## Install Redis

1. Install redis with docker

```bash
docker run -d --name redis -p 6379:6379 redis
```

2. Verify if redis is running

```bash
docker exec -it redis redis-cli
ping
```

If PONG is returned, redis is running.

## IMPORTANT!

The followings are required to be setup in order for the system to fully work as whole

1. Environment for deployment:

- DB_USERNAME
- DB_PASSWORD
- JWT_SECRET_KEY
- AWS_S3_BUCKET_NAME
- AWS_REGION
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- FIREBASE_PRIVATE_KEY_ID

2. keys in system_data table for "LHDN e-invoice":

- erp_client_id
- erp_client_secret (encrypted)
- main_taxpayer_tin

3. Keys to be given to Grabfood Developer for "GRABFOOD" and on system_data as well:

- GrabFoodPartnerClientID
- GrabFoodPartnerClientSecret (encrypted)

4. Keys to setup in system for "GRABFOOD"

- grabFoodClientID
- grabFoodClientSecret (encrypted)

5. Need to setup fiuu in the future config in the future..

- Offline config done
- Online config not
