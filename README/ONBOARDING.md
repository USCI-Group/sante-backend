# Onboarding Steps

Welcome to project Sante! Follow these steps to get your development environment up and running.

## Prerequisites

Before you begin, make sure you have the following installed:

- [Go](https://go.dev/doc/install) (version 1.20 or higher recommended)
- Install [Encore](https://encore.dev/docs/install) (backend framework)
- [Docker](https://www.docker.com/get-started)
- [Git](https://git-scm.com/downloads)
- [pgAdmin](https://www.pgadmin.org/download/) (for managing and inspecting the PostgreSQL database, optional but recommended)
- Install Node Version Manager (nvm)
- Install Node v24
- Install atlas (for schema migration tool)
- Install redis

### Install nvm and node

1. **Install nvm**  
   Run the following command in your terminal:

   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
   ```

2. **Install Node.js (v24)**  
   ```bash
   nvm install 24
   ```

3. **Verify installation**

   ```bash
   node -v
   npm -v
   ```

### Install Atlas
   ```bash
   curl -sSf https://atlasgo.sh | sh
   ```

### Install Redis

1. Install redis with docker
```bash
docker run -d --name redis -p 6379:6379 redis
```

## Set up Sante Backend (encore.go)

### clone project

1. **Sante Backend**
   ```bash
    git clone https://github.com/AFED-digital/sante-backend.git
   ```

2. **Sante Admin Web**
   ```bash
   git clone https://github.com/AFED-digital/sante-admin-web.git
   ```

Now you're set. To run the backend, proceed to this readme https://github.com/AFED-digital/sante-backend/blob/3aeadbc8f8bf3c815eb973cde0f9508f1545348d/README_CONFIG.md



