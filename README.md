# 🚀 URX — Universal Runtime

Run applications anywhere — package once, execute anywhere.

URX is a CLI tool that packages applications into portable `.urx` artifacts and executes them using containers — without requiring Dockerfiles or complex setup.

---

## ⚙️ Installation

```bash
git clone https://github.com/vijaythakur89/urx.git
cd urx
go build -o urx ./cmd/urx-cli
sudo mv ./urx /usr/local/bin/
```

---

## ⚡ Quick Start

```bash
urx build demo
urx run app.urx
urx ps #to get the run ID.
urx logs <run-id>
```

---

## 📺 Example Output

```bash
urx run app.urx
```

```
[URX] Running container: urx-xxxx
[URX] View logs: urx logs urx-xxxx
```

```bash
urx logs urx-xxxx
```

```
🚀 URX app started
running...
running...
```

---

---

## ⚙️ Installation

> Requires Docker to be installed and running.

```bash
git clone https://github.com/vijaythakur89/urx.git
cd urx
go build -o urx ./cmd/urx-cli
sudo mv ./urx /usr/local/bin/

## 📁 Example (demo app)

```
demo/
├── app.py
└── manifest.yaml
```

### manifest.yaml

```yaml
runtime: python
entrypoint: app.py
```

---

## 🧠 How it works

```
source code → urx build → .urx artifact → urx run → container execution
```

---

## 🛠 Commands

* `urx build`   → Build artifact
* `urx run`     → Run application
* `urx ps`      → List runs with status
* `urx logs`    → Show logs (`-f` to follow)
* `urx stop`    → Stop container
* `urx rm`      → Remove container + metadata
* `urx inspect` → Inspect artifact
* `urx version` → Show version

---

## 🏗 Architecture

* CLI (Cobra)
* Artifact builder (tar + manifest)
* Runtime engine (Docker-based execution)
* Metadata storage (`~/.urx`)

---

## 🚧 Roadmap

* Kubernetes integration
* Multi-runtime support
* Isolation modes
* Remote execution

---

## 📌 Why URX?

Running apps today requires:
- Writing Dockerfiles
- Managing environments
- Handling runtime inconsistencies

URX abstracts all of this into a single workflow.

---

## ⚠️ Status

This is an early-stage project (MVP). Contributions and feedback are welcome.

---
## 🏗 Architecture

- CLI (Cobra)
- Artifact builder (tar + manifest)
- Runtime engine (Docker)
- Metadata store (~/.urx)
- --
## 📜 License

MIT
