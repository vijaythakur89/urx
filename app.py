import time

print("🚀 URX app started")

# read from mounted volume
with open("/app/data/input.txt", "r") as f:
    content = f.read()

print("Read from volume:", content)

# write back to host via volume
with open("/app/data/output.txt", "w") as f:
    f.write("written-by-urx")

while True:
    print("running...")
    time.sleep(2)
