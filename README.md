# procwatch

> Minimal daemon that monitors process resource usage and emits structured logs or alerts on thresholds.

---

## Installation

```bash
pip install procwatch
```

Or install from source:

```bash
git clone https://github.com/yourname/procwatch.git && cd procwatch && pip install .
```

---

## Usage

Start watching a process by name or PID:

```bash
procwatch --pid 1234 --cpu 80 --mem 512
```

```bash
procwatch --name nginx --cpu 70 --mem 256 --interval 5
```

**Example structured log output:**

```json
{"timestamp": "2024-05-10T14:32:01Z", "pid": 1234, "name": "nginx", "cpu_percent": 73.2, "mem_mb": 198.4, "alert": false}
{"timestamp": "2024-05-10T14:32:06Z", "pid": 1234, "name": "nginx", "cpu_percent": 81.5, "mem_mb": 201.1, "alert": true, "reason": "cpu_threshold_exceeded"}
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--pid` | Target process ID | — |
| `--name` | Target process name | — |
| `--cpu` | CPU usage alert threshold (%) | `90` |
| `--mem` | Memory alert threshold (MB) | `1024` |
| `--interval` | Poll interval in seconds | `10` |
| `--log` | Log output file path | stdout |
| `--exit-on-alert` | Stop watching after the first alert is triggered | `false` |

---

## Requirements

- Python 3.8+
- [`psutil`](https://github.com/giampaolo/psutil)

---

## License

MIT © 2024 yourname
