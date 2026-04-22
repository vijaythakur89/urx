from http.server import HTTPServer, BaseHTTPRequestHandler

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write("Hello from URX".encode())

print("URX app started")

# health indicator
with open("/tmp/urx_health", "w") as f:
    f.write("ok")

server = HTTPServer(("0.0.0.0", 8080), Handler)
server.serve_forever()
