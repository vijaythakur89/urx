# URX System Architecture

**Project:** Universal Runtime eXecutor (URX)

**Document Version:** 0.1

**Project Version:** 0.3.x

**Author:** Vijay Kumar

**Status:** Active Development

---

# Document Revision History

| Version | Date | Description | Author |
|----------|------|-------------|--------|
| 0.1 | 2026 | Initial system architecture documentation | Vijay Thakur |

---

# Table of Contents

1. Purpose
2. Architecture Principles
3. Design Goals
4. Architectural Layers
5. High-Level Architecture
6. Major Components
7. Repository Layout
8. Core Workflows
9. Build Lifecycle
10. Runtime Lifecycle
11. Deployment Lifecycle
12. Metadata Lifecycle
13. Runtime Metadata Structure
14. Runtime Data Flow
15. Architecture Decisions
16. Current Runtime Architecture
17. Target Architecture
18. Current Limitations
19. Conclusion

---

# Purpose

This document describes the internal architecture of the Universal Runtime eXecutor (URX).

It explains how the major software components interact, how application artifacts are built and executed, how runtime metadata is managed, and how the overall system has been designed to support future runtime backends.

This document serves as the primary architectural reference for contributors and maintainers.

---

# Architecture Principles

The URX architecture has been designed around the following principles:

## Simplicity

The runtime should remain lightweight and easy to understand.

## Separation of Responsibilities

Each package should have a clearly defined responsibility.

Examples:

- CLI handles user interaction.
- Builder packages applications.
- Runtime executes applications.
- Storage manages metadata.

## Runtime Independence

The application packaging format should not depend on the runtime implementation.

Today URX uses Docker.

Future implementations may use:

- containerd
- Firecracker
- Kubernetes
- Remote execution environments

without changing the artifact format.

## Extensibility

New features should be added by extending existing components rather than modifying unrelated modules.

## Maintainability

The codebase should remain readable, testable, and modular.

# Design Goals

The architecture of URX has been designed to satisfy the following engineering goals.

## Modularity

Components should have clearly defined responsibilities.

## Portability

Artifacts should remain portable across supported runtime implementations.

## Runtime Independence

Packaging should not depend on Docker or any specific execution backend.

## Testability

Core functionality should be testable without requiring a full runtime environment.

## Extensibility

New runtime implementations should be added with minimal impact on existing packages.

## Observability

Runtime state should be discoverable through metadata and CLI inspection commands.

---

# Architectural Layers

The URX platform is organized into logical layers.

Each layer has a single responsibility and communicates only with adjacent layers.

```text
+------------------------------------------------------+
|                  User Layer                          |
|               CLI Commands (Cobra)                   |
+------------------------------------------------------+
                         │
                         ▼
+------------------------------------------------------+
|              Application Layer                       |
|      Build • Deploy • Run • Status • Logs           |
+------------------------------------------------------+
                         │
                         ▼
+------------------------------------------------------+
|               Runtime Layer                          |
|      Docker Runtime (Current Implementation)         |
+------------------------------------------------------+
                         │
                         ▼
+------------------------------------------------------+
|               Storage Layer                          |
|        Metadata • Logs • Runtime State              |
+------------------------------------------------------+
```

Each layer can evolve independently provided its public interfaces remain stable.

This layered architecture improves modularity, testability, and long-term maintainability.

---

# High-Level Architecture

The following diagram illustrates the major software components within URX.

```text
                   +----------------------------+
                   |         User CLI           |
                   | build run deploy ps logs   |
                   +-------------+--------------+
                                 |
                                 v
                   +----------------------------+
                   |        Cobra Commands      |
                   +-------------+--------------+
                                 |
                                 v
                   +----------------------------+
                   |     Runtime Controller     |
                   +-------------+--------------+
                                 |
             +-------------------+-------------------+
             |                                       |
             v                                       v
   +----------------------+             +----------------------+
   | Artifact Builder     |             | Artifact Extractor   |
   +----------------------+             +----------------------+
             |                                       |
             +-------------------+-------------------+
                                 |
                                 v
                   +----------------------------+
                   |     Runtime Engine         |
                   |      (Docker Today)        |
                   +-------------+--------------+
                                 |
                                 v
                   +----------------------------+
                   |   Running Application      |
                   +-------------+--------------+
                                 |
                                 v
                   +----------------------------+
                   | Metadata Storage           |
                   | ~/.urx/runs               |
                   +----------------------------+
```

# Major Components

## CLI Layer

Location:

```
cmd/urx-cli/
```

Responsibilities:

- Parse user commands.
- Validate user input.
- Dispatch requests to the appropriate package.
- Present results to the user.

The CLI should contain minimal business logic.

---

## Artifact Layer

Location:

```
artifacts/
```

Responsibilities:

- Build `.urx` artifacts.
- Read manifests.
- Inspect packaged artifacts.

---

## Runtime Layer

Location:

```
runtime/local/
```

Responsibilities:

- Extract artifacts.
- Configure runtime execution.
- Launch containers.
- Configure networking.
- Configure volumes.
- Configure environment variables.
- Deploy long-running services.

---

## Storage Layer

Location:

```
pkg/storage/
```

Responsibilities:

- Store runtime metadata.
- Load metadata.
- Manage runtime directories.
- Persist execution information.

Runtime metadata is intentionally separated from Docker state.

---

# Repository Layout

```text
URX
│
├── cmd/
│      CLI commands
│
├── artifacts/
│      Packaging engine
│
├── runtime/
│      Runtime implementation
│
├── pkg/
│      Shared libraries
│
├── demo/
│      Reference application
│
├── docs/
│      Project documentation
│
└── .github/
       CI/CD workflows
```

Each directory has a single primary responsibility.

This organization helps keep the project modular and maintainable.

---

# Core Workflows

URX provides three primary workflows that together form the application lifecycle.

## Build Workflow

The build workflow packages an application into a portable `.urx` artifact.

```
Source Application
        │
        ▼
   urx build
        │
        ▼
 Manifest Processing
        │
        ▼
 Package Files
        │
        ▼
  Generate app.urx
```

Output:

- Portable `.urx` artifact
- Embedded manifest
- Application source

---

## Run Workflow

The run workflow executes a packaged application as a temporary container.

```
app.urx
    │
    ▼
Extract Artifact
    │
    ▼
Read Manifest
    │
    ▼
Configure Runtime
    │
    ▼
Launch Container
```

Characteristics:

- One-time execution
- No restart policy
- Suitable for testing and development

---

## Deploy Workflow

The deploy workflow runs the application as a managed service.

```
app.urx
    │
    ▼
Extract Artifact
    │
    ▼
Configure Runtime
    │
    ▼
Create Service
    │
    ▼
Persist Metadata
    │
    ▼
Display Service URL
```

Additional deployment features include:

- Restart policy
- Port mapping
- Runtime metadata
- Service discovery

---

# Build Lifecycle

The build process converts a source application into a portable `.urx` artifact.

The artifact contains:

* Application source code
* Manifest
* Runtime metadata required for execution

The build process is intentionally independent of the runtime implementation.

## Build Workflow

```text
Developer
     │
     ▼
urx build demo/
     │
     ▼
Read manifest
     │
     ▼
Package source files
     │
     ▼
Generate .urx artifact
     │
     ▼
Save artifact
```

## Sequence

1. The user executes the `urx build` command.
2. The builder validates the project directory.
3. A manifest is generated or loaded.
4. Source files are packaged into a TAR archive.
5. The manifest is embedded into the archive.
6. The archive is written as a `.urx` artifact.

The resulting artifact is runtime-independent and can be executed by any supported backend.

# Runtime Lifecycle

The runtime is responsible for executing a packaged application.

Current implementation:

```
Docker
```

Future implementations may include:

* containerd
* Firecracker
* Kubernetes

## Runtime Flow

```text
User

↓

urx run app.urx

↓

Extract Artifact

↓

Load Manifest

↓

Configure Runtime

↓

Launch Container

↓

Application Running
```

The runtime implementation is isolated from the packaging logic.

This separation allows additional runtime backends to be introduced without modifying the artifact format.


# Deployment Lifecycle

Deployment mode extends runtime execution by configuring the application as a long-running service.

Additional deployment responsibilities include:

* Restart policy
* Port mapping
* Volume mounting
* Environment variable injection
* Runtime metadata persistence

## Deployment Flow

```text
urx deploy

↓

Extract Artifact

↓

Read Manifest

↓

Allocate Host Port

↓

Configure Docker

↓

Launch Service

↓

Persist Metadata

↓

Display Service URL
```

Deployment mode is designed for lightweight service hosting rather than orchestration.

# Metadata Lifecycle

URX maintains runtime metadata independently of Docker.

Metadata is stored under:

```
~/.urx/runs/
```

Each execution generates a dedicated runtime directory.

Example:

```text
~/.urx
└── runs
    └── urx-xxxxxxxx
        └── meta.json
```

The metadata currently includes:

* Container identifier
* Artifact name
* Execution timestamp
* Assigned service port

This metadata enables runtime inspection without relying exclusively on Docker APIs.

---

# Runtime Metadata Structure

URX maintains its own runtime metadata independent of Docker.

Current directory layout:

```text
~/.urx
│
└── runs
    │
    └── urx-xxxxxxxx
          │
          ├── meta.json
          └── logs.txt (future)
```

Example metadata:

```json
{
  "id": "urx-b3d953c1c285e70b",
  "artifact": "app.urx",
  "timestamp": "2026-05-05T23:09:22+05:30",
  "port": 8080
}
```

This metadata allows URX to provide runtime information without depending exclusively on Docker APIs.

---

# Runtime Data Flow

The following diagram illustrates how data flows through the system during deployment.

```text
Application Source

        │

        ▼

Builder

        │

        ▼

.app.urx Artifact

        │

        ▼

Runtime

        │

        ▼

Container

        │

        ├──────────► Logs

        ├──────────► Health

        └──────────► Metadata

                     │

                     ▼

              ~/.urx/runs
```

The runtime produces operational information that can be queried through CLI commands such as:

* `urx ps`
* `urx status`
* `urx logs`

# Architecture Decisions

The following architectural decisions have guided the implementation of URX.

## ADR-001

### Decision

Use TAR as the artifact format.

### Rationale

TAR is portable, simple, and well-supported across platforms.

---

## ADR-002

### Decision

Use Docker as the initial runtime backend.

### Rationale

Docker provides a mature runtime while the overall URX architecture is developed.

The runtime implementation will later be abstracted to support additional backends.

---

## ADR-003

### Decision

Separate runtime metadata from Docker state.

### Rationale

URX should maintain its own runtime information independent of any specific container runtime.

---

## ADR-004

### Decision

Keep CLI commands lightweight.

### Rationale

Business logic belongs inside reusable packages rather than the command layer.

This improves maintainability and testability.

# Current Runtime Architecture

The current implementation uses Docker as the execution backend.

This dependency is intentional and temporary.

The long-term architecture introduces a runtime abstraction layer.

Future architecture:

```text
                URX Runtime

                     │

     ┌───────────────┼───────────────┐

     ▼               ▼               ▼

 Docker        containerd      Firecracker

                                     │

                                     ▼

                              Kubernetes
```

Application artifacts will remain unchanged regardless of the selected runtime backend.

Only the runtime implementation will vary.

---

# Target Architecture

The long-term architecture introduces a runtime abstraction layer between URX and the execution backend.

```text
                     User CLI
                         │
                         ▼
                   URX Core Engine
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
     Docker Runtime  containerd   Firecracker
          │              │              │
          └──────────────┼──────────────┘
                         ▼
                 Runtime Interface
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
      Local Host     Kubernetes     Remote Host
```

This architecture allows new runtimes to be introduced without changing the application packaging format or user-facing CLI.

---

# Current Limitations

The current implementation intentionally focuses on establishing a stable engineering foundation.

Known limitations include:

- Docker is the only supported runtime backend.
- Secrets management is not yet implemented.
- Runtime plugins are not yet available.
- Kubernetes deployment is not supported.
- Remote execution is not supported.
- Distributed scheduling is outside the current scope.

These limitations are tracked as part of the long-term project roadmap.

---

# Conclusion

The URX architecture emphasizes modularity, portability, and long-term extensibility.

By separating packaging, runtime execution, metadata management, and command-line interaction into distinct components, the project provides a clean foundation for future growth.

The current implementation focuses on establishing a stable engineering platform before introducing advanced runtime capabilities such as runtime abstraction, alternative execution backends, and distributed deployment models.

This architecture enables URX to evolve incrementally while preserving compatibility with existing artifacts and workflows.

---

# References

Related documentation:

- 01_Project_Overview.md
- 03_Admin_Setup_Guide.md (Planned)
- 04_Developer_Guide.md (Planned)
- 05_Artifact_Format.md (Planned)
- 06_Runtime_Architecture.md (Planned)
- 07_Testing_and_CI.md (Planned)

---
