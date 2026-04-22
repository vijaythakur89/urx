# 🚀 URX — Universal Runtime

Run applications anywhere with a simple CLI.

URX packages applications into portable `.urx` artifacts and executes them using containers — without requiring Dockerfiles or complex setup.

---

## ⚡ Quick Start

```bash
urx build demo
urx run app.urx
urx ps
urx logs <run-id>
```

---

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
* `urx ps`      → List runs
* `urx logs`    → Show logs (`-f` to follow)
* `urx stop`    → Stop container
* `urx rm`      → Remove container
* `urx inspect` → Inspect artifact
* `urx version` → Show version

---

## 🚧 Roadmap

* Kubernetes integration
* Multi-runtime support
* Isolation modes
* Remote execution

---

## 📌 Why URX?

Modern application execution is complex:

* Dockerfiles
* Environment setup
* Runtime inconsistencies

URX simplifies this into a single workflow.

---

## 📜 License

MIT

