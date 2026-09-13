import hashlib
import json
import os
from collections import defaultdict
from pathlib import Path
from urllib.parse import parse_qsl, urlencode, urlsplit

from mitmproxy import http


class DawgCapture:
    def __init__(self):
        self.output_dir = Path(os.environ["DAWG_SESSION_DIR"])
        hosts = json.loads(os.environ.get("DAWG_INTERNAL_HOSTS", "[]"))
        self.internal_hosts = set(hosts) if hosts else set()
        self.ordinals = defaultdict(int)

    def running(self):
        print(json.dumps({"status": "capturing"}), flush=True)

    def response(self, flow: http.HTTPFlow):
        pair = {
            "id": flow.id,
            "timestamp": flow.response.timestamp_end,
            "request": {
                "method": flow.request.method,
                "url": flow.request.pretty_url,
                "headers": dict(flow.request.headers),
                "body": flow.request.get_text(strict=False),
            },
            "response": {
                "status": flow.response.status_code,
                "headers": dict(flow.response.headers),
                "body": flow.response.get_text(strict=False),
            },
            "direction": "backend-to-external",
            "durationMs": int((flow.response.timestamp_end - flow.request.timestamp_start) * 1000),
        }
        target = self.output_dir / "http" / "backend.jsonl"
        if flow.request.host not in self.internal_hosts:
            match_key = self.match_key(flow)
            pair["matchKey"] = match_key
            target = self.output_dir / "cassettes" / "thirdparty.jsonl"
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("a", encoding="utf-8") as output:
            output.write(json.dumps(pair, sort_keys=True) + "\n")

    def match_key(self, flow: http.HTTPFlow):
        parsed = urlsplit(flow.request.pretty_url)
        normalized_query = urlencode(sorted(parse_qsl(parsed.query, keep_blank_values=True)))
        path = parsed.path + ("?" + normalized_query if normalized_query else "")
        body = self.canonical_body(flow.request.get_text(strict=False))
        fingerprint = "sha256:" + hashlib.sha256(body.encode("utf-8")).hexdigest()
        base = (flow.request.method, path, fingerprint)
        self.ordinals[base] += 1
        return {
            "method": flow.request.method,
            "pathPattern": path,
            "bodyFingerprint": fingerprint,
            "ordinal": self.ordinals[base],
        }

    @staticmethod
    def canonical_body(body):
        try:
            return json.dumps(json.loads(body), sort_keys=True, separators=(",", ":"))
        except (TypeError, ValueError):
            return body


addons = [DawgCapture()]