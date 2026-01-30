import csv
import matplotlib.pyplot as plt

path = r"logs\bg_metrics.csv"
steps = []
vals = []

with open(path, newline="", encoding="utf-8") as f:
    r = csv.DictReader(f)
    i = 0
    for row in r:
        if row["metric"] != "success_rate":
            continue
        i += 1
        steps.append(i)
        vals.append(float(row["value"]))

plt.figure()
plt.plot(steps, vals, marker="o")
plt.xlabel("sample")
plt.ylabel("success_rate (%)")
plt.title("goruut backtest success rate over time")
plt.grid(True)
plt.tight_layout()
plt.savefig(r"logs\bg_success_rate.png", dpi=150)
print("Wrote logs\\bg_success_rate.png")
