# ⚡ Workflow Editor

A modern workflow editor app for designing and executing custom automation workflows (e.g., weather notifications). Users can visually build workflows, configure parameters, and view real-time execution results.

## Design/Tooling decisions
+ golang-migrate - went with this tool for db migrations based on having the most stars of similar tooling in go-land and still being maintained with occasional updates (last being Apr 24, 2 months ago as of writing)
  + I was burning some time figuring out the golang db migration tool landscape, and decided to stop trying to find a more type-strict go-native tool where you could write migrations using a query-builder of sorts, that could use go-defined types and structs
+ table design
  + nodes/edges tables with the right indexes, and recursive CTEs, can effectively operates like a graph DB
  + however, I opted to just store the workflow_id of each node against each node/edge to avoid those potentially expensive operations for sufficiently large workflows
    + this assumes a 1:1 relationship between workflows, and nodes/edges
    + and as such precludes re-use of nodes/edges across workflows in future - making this call for now for simplicity, and to not prematurely optimise for a use case that hasn't been defined in the project (take-home task) spec
  + dumping all of `data` into just a jsonb column, as based on the task right now and even assuming new nodes types, etc. will be introduced in future, we shouldn't need to search on data within the data block at least for the purpose of the workflow editor
    + there may be a use case for searching over them more for analytics/usage purposes, but that should be deferred to another DB better served for that use case, e.g. elastic search or similar

## Assumptions
+ the POST /execute endpoint is supposed to provide the full workflow definition + form inputs - from the initial clone, it seems that it only provides form inputs and not the full workflow (nodes+edges data), assuming this is part of the task
  + also because otherwise it won't be possible to update the workflow as it exists in the frontend, as part of the POST /execute call
+ I took existing code to be the source of truth, where there were minor conflicts with the spec:
  + for execute endpoint:
    + spec: Execute Endpoint (POST /execute)
    + code: /workflows/{id}/execute
  + for get workflow endpoint:
    + spec: GET /workflow/{id}
    + code: GET /workflows/{id}


## 🛠️ Tech Stack

- **Frontend:** React + TypeScript, @xyflow/react (drag-and-drop), Radix UI, Tailwind CSS, Vite
- **Backend:** Go API, PostgreSQL database
- **DevOps:** Docker Compose for orchestration, hot reloading for rapid development

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose (recommended for development)
- Node.js v18+ (for local frontend development)
- Go v1.23+ (for local backend development)

> **Tip:** Node.js and Go are only required if you want to run frontend or backend outside Docker.

### 1. Start All Services

```bash
docker-compose up --build
```

- This launches frontend, backend, and database with hot reloading enabled for code changes.
- To stop and clean up:
  ```bash
  docker-compose down
  ```

### 2. Access Applications

- **Frontend (Workflow Editor):** [http://localhost:3003](http://localhost:3003)
- **Backend API:** [http://localhost:8086](http://localhost:8086)
- **Database:** PostgreSQL on `localhost:5876`

### 3. Verify Setup

1. Open [http://localhost:3003](http://localhost:3003) in your browser.
2. You should see the workflow editor with sample nodes.

## 🏗️ Project Architecture

```text
workflow-code-test/
├── api/                    # Go Backend (Port 8086)
│   ├── main.go
│   ├── services/
│   ├── pkg/
│   ├── go.mod
│   └── Dockerfile
├── web/                    # React Frontend (Port 3003)
│   ├── src/
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## 🔧 Development Workflow

### 🌐 Frontend

- Edit files in `web/src/` and see changes instantly at [http://localhost:3003](http://localhost:3003) (hot reloading via Vite).

### 🖥️ Backend

- Edit files in `api/` and changes are reflected automatically (hot reloading in Docker).
- If you add new dependencies or make significant changes, rebuild the API container:
  ```bash
  docker-compose up --build api
  ```

### 🗄️ Database

- Schema/configuration details: see [API README](api/README.md#database)
- After schema changes or migrations, restart the database:
  ```bash
  docker-compose restart postgres
  ```
- To apply schema changes to the API after updating the database:
  ```bash
  docker-compose restart api
  ```