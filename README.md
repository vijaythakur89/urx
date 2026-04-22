# 🚀 URX — Universal Runtime eXecutor

URX is a lightweight CLI tool that lets you package and run applications as portable `.urx` artifacts.

It simplifies application execution by removing the need for Dockerfiles and complex environment setup.

---

## ✨ Features

* 📦 Package apps into a single `.urx` artifact
* ▶️ Run apps locally with a single command
* 🚀 Deploy apps as long-running services
* 📁 Volume mounting support
* 🔐 Environment variable injection (`.env`, CLI, system env)
* ❤️ Health checks for running apps
* 🐳 Custom base image support
* 🌐 Service port exposure
* 🔍 Inspect running apps (`ps`, `status`, `logs`)

---

## ⚡ Installation

```bash
git clone https://github.com/vijaythakur89/urx.git
cd urx
go build -o urx ./cmd/urx-cli
sudo mv ./urx /usr/local/bin/

---

## 🧩 Requirements

- Docker must be installed and running
- Go (only required for building URX from source)
> URX relies on Docker as the container runtime (for now).

---

## ⚡ Usage

### 1️⃣ Build artifact

```bash
urx build demo
```

---

### 2️⃣ Run (one-off execution)

```bash
urx run app.urx
```

---

### 3️⃣ Deploy (service mode)

```bash
urx deploy app.urx
```

---

### 4️⃣ List running apps

```bash
urx ps
```

---

### 5️⃣ Inspect a running app

```bash
urx status <id>
```

---

### 6️⃣ View logs

```bash
urx logs <id>
```

---

## 📄 Example Project

```
demo/
├── app.py
├── manifest.yaml
└── .env
```

---

## 📄 Example manifest.yaml

```yaml
name: demo
runtime: python
base_image: python:3.11-slim
entrypoint: app.py
isolation: low

port: 8080

volumes:
  - "/home/user/data:/app/data"

env:
  - TEST
```

---

## 🔐 Environment Variables

URX supports environment variables from multiple sources:

1. `.env` file (recommended)
2. System environment
3. CLI flags (future support)

### Example `.env`

```
TEST=hello
API_KEY=xyz
```

---

## 🌐 Example HTTP App

```python
from http.server import HTTPServer, BaseHTTPRequestHandler

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write("Hello from URX 🚀".encode())

# health indicator
with open("/tmp/urx_health", "w") as f:
    f.write("ok")

server = HTTPServer(("0.0.0.0", 8080), Handler)
server.serve_forever()
```

---

## 🌍 Accessing Deployed App

After deployment:

```bash
urx deploy app.urx
```

Access your app:

```
http://localhost:8080
```

---

## 🧠 How it works

```
source code → urx build → .urx artifact → urx run/deploy → container execution
```

---

## 🛠 Commands

| Command       | Description       |
| ------------- | ----------------- |
| `urx build`   | Build artifact    |
| `urx run`     | Run application   |
| `urx deploy`  | Run as service    |
| `urx ps`      | List running apps |
| `urx status`  | Inspect app       |
| `urx logs`    | View logs         |
| `urx stop`    | Stop container    |
| `urx rm`      | Remove container  |
| `urx inspect` | Inspect artifact  |
| `urx version` | Show version      |

---

## 📌 Why URX?

Modern application execution is complex:

* Dockerfiles
* Environment setup
* Runtime inconsistencies

URX simplifies this into a single workflow:

👉 **Build once, run anywhere**

---

## 🚧 Roadmap

* Auto port allocation + URL output
* Kubernetes integration
* Multi-runtime support (Node, Go, Java)
* Remote execution
* Secrets management enhancements

---

## 🤝 Contributing

Contributions, ideas, and feedback are welcome!

---

## 📄 License

MIT License

