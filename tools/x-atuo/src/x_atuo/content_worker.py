"""Supervisor for the combined Sub2API content worker."""

from __future__ import annotations

import os
import signal
import subprocess
import sys
import time
from dataclasses import dataclass


@dataclass(frozen=True)
class ChildSpec:
    name: str
    args: list[str]


def _children() -> list[ChildSpec]:
    return [
        ChildSpec("x-auto", ["uvicorn", "x_atuo.automation.api:app", "--host", "0.0.0.0", "--port", os.getenv("X_AUTO_PORT", "8000")]),
        ChildSpec("hot-rss", [sys.executable, "-m", "x_atuo.hot_rss_worker"]),
    ]


def main() -> int:
    if "--healthcheck" in sys.argv:
        checks = [
            [sys.executable, "-m", "x_atuo.hot_rss_worker", "--healthcheck"],
            [
                sys.executable,
                "-c",
                "import urllib.request; urllib.request.urlopen('http://127.0.0.1:%s/healthz', timeout=5).read()" % os.getenv("X_AUTO_PORT", "8000"),
            ],
        ]
        for command in checks:
            result = subprocess.run(command, env=os.environ)
            if result.returncode != 0:
                return result.returncode
        print("[content-worker] healthcheck ok", flush=True)
        return 0

    processes: dict[str, subprocess.Popen[bytes]] = {}
    shutting_down = False

    def stop_all(signum: int = signal.SIGTERM, _frame: object | None = None) -> None:
        nonlocal shutting_down
        shutting_down = True
        for process in processes.values():
            if process.poll() is None:
                process.send_signal(signum)

    signal.signal(signal.SIGTERM, stop_all)
    signal.signal(signal.SIGINT, stop_all)

    for child in _children():
        process = subprocess.Popen(child.args, env=os.environ)
        processes[child.name] = process
        print(f"[content-worker] started {child.name} pid={process.pid}", flush=True)

    exit_code = 0
    try:
        while processes:
            for name, process in list(processes.items()):
                code = process.poll()
                if code is None:
                    continue
                processes.pop(name, None)
                print(f"[content-worker] {name} exited code={code}", flush=True)
                if not shutting_down:
                    exit_code = code or 1
                    stop_all()
            if processes:
                time.sleep(1)
    finally:
        stop_all()
        for process in processes.values():
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
